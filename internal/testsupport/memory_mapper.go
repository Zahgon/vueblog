package testsupport

import (
	"sort"
	"sync"

	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// MemoryBlogMapper is an in-memory mapper.BlogMapper with MySQL-equivalent semantics.
// The Java suite could only ever run against a live MySQL instance; this lets
// the migrated tests exercise the same code paths deterministically.
type MemoryBlogMapper struct {
	mu     sync.Mutex
	rows   map[int64]*entity.Blog
	nextId int64
}

func NewMemoryBlogMapper() *MemoryBlogMapper {
	return &MemoryBlogMapper{rows: map[int64]*entity.Blog{}, nextId: 1}
}

// Seed inserts a row preserving its id, mirroring the INSERT statements in
// vueblog.sql.
func (m *MemoryBlogMapper) Seed(b *entity.Blog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *b
	m.rows[*b.Id] = &cp
	if *b.Id >= m.nextId {
		m.nextId = *b.Id + 1
	}
}

func (m *MemoryBlogMapper) SelectById(id int64) (*entity.Blog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.rows[id]
	if !ok {
		return nil, nil // MyBatis returns null for a missing row
	}
	cp := *b
	return &cp, nil
}

func (m *MemoryBlogMapper) SelectPage(page *mybatisplus.Page[*entity.Blog], qw *mybatisplus.QueryWrapper) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	all := make([]*entity.Blog, 0, len(m.rows))
	for _, b := range m.rows {
		cp := *b
		all = append(all, &cp)
	}

	// ORDER BY created DESC. MySQL's sort is not stable across equal keys, but
	// the ids are unique here so we tie-break on id descending for determinism.
	for _, o := range qw.Orders() {
		if o == "created DESC" {
			sort.SliceStable(all, func(i, j int) bool {
				ti, tj := all[i].Created, all[j].Created
				if ti == nil || tj == nil {
					return tj == nil && ti != nil
				}
				if ti.T.Equal(tj.T) {
					return *all[i].Id > *all[j].Id
				}
				return ti.T.After(tj.T)
			})
		}
	}

	page.Total = int64(len(all))
	page.ComputePages()

	offset := page.Offset()
	if offset >= int64(len(all)) {
		page.Records = []*entity.Blog{}
		return nil
	}
	end := offset + page.Size
	if end > int64(len(all)) {
		end = int64(len(all))
	}
	page.Records = all[offset:end]
	return nil
}

func (m *MemoryBlogMapper) Insert(b *entity.Blog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextId
	m.nextId++
	b.Id = &id // MP writes the generated key back onto the entity (IdType.AUTO)
	cp := *b
	m.rows[id] = &cp
	return nil
}

func (m *MemoryBlogMapper) UpdateById(b *entity.Blog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b.Id == nil {
		return nil
	}
	cp := *b
	m.rows[*b.Id] = &cp
	return nil
}

// MemoryUserMapper is an in-memory UserMapper.
type MemoryUserMapper struct {
	mu   sync.Mutex
	rows map[int64]*entity.User
}

func NewMemoryUserMapper() *MemoryUserMapper {
	return &MemoryUserMapper{rows: map[int64]*entity.User{}}
}

func (m *MemoryUserMapper) Seed(u *entity.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *u
	m.rows[*u.Id] = &cp
}

func (m *MemoryUserMapper) SelectById(id int64) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.rows[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *MemoryUserMapper) SelectOne(qw *mybatisplus.QueryWrapper) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cols, vals := qw.Eqs()
	for _, u := range m.rows {
		match := true
		for i, c := range cols {
			if c == "username" {
				if s, ok := vals[i].(string); !ok || u.Username != s {
					match = false
				}
			}
		}
		if match {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}
