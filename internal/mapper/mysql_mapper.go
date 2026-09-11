package mapper

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// blogColumns matches the m_blog DDL in vueblog.sql. MP derives these from the
// entity field names via its default camel-to-underscore strategy.
const blogColumns = "id, user_id, title, description, content, created, status"

// MySQLBlogMapper is the production BlogMapper, equivalent to what MyBatis-Plus
// generates for BaseMapper<Blog>.
type MySQLBlogMapper struct{ DB *sql.DB }

func NewMySQLBlogMapper(db *sql.DB) *MySQLBlogMapper { return &MySQLBlogMapper{DB: db} }

func scanBlog(row interface{ Scan(...any) error }) (*entity.Blog, error) {
	var (
		b       entity.Blog
		id      sql.NullInt64
		userId  sql.NullInt64
		created sql.NullTime
		status  sql.NullInt32
		content sql.NullString
	)
	if err := row.Scan(&id, &userId, &b.Title, &b.Description, &content, &created, &status); err != nil {
		return nil, err
	}
	if id.Valid {
		v := id.Int64
		b.Id = &v
	}
	if userId.Valid {
		v := userId.Int64
		b.UserId = &v
	}
	if content.Valid {
		b.Content = content.String
	}
	if created.Valid {
		b.Created = lang.NewLocalDate(created.Time)
	}
	if status.Valid {
		v := status.Int32
		b.Status = &v
	}
	return &b, nil
}

func (m *MySQLBlogMapper) SelectById(id int64) (*entity.Blog, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", blogColumns, entity.BlogTableName)
	b, err := scanBlog(m.DB.QueryRow(q, id))
	if err == sql.ErrNoRows {
		return nil, nil // MyBatis returns null, not an error
	}
	return b, err
}

func (m *MySQLBlogMapper) SelectPage(page *mybatisplus.Page[*entity.Blog], qw *mybatisplus.QueryWrapper) error {
	where, args := qw.WhereSQL()

	// PaginationInterceptor issues a COUNT first when isSearchCount is true.
	if page.SearchCount {
		countQ := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", entity.BlogTableName, where)
		if err := m.DB.QueryRow(countQ, args...).Scan(&page.Total); err != nil {
			return err
		}
	}
	page.ComputePages()

	q := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT ? OFFSET ?",
		blogColumns, entity.BlogTableName, where, qw.OrderSQL())
	rows, err := m.DB.Query(q, append(append([]any{}, args...), page.Size, page.Offset())...)
	if err != nil {
		return err
	}
	defer rows.Close()

	records := []*entity.Blog{}
	for rows.Next() {
		b, err := scanBlog(rows)
		if err != nil {
			return err
		}
		records = append(records, b)
	}
	page.Records = records
	return rows.Err()
}

func (m *MySQLBlogMapper) Insert(b *entity.Blog) error {
	q := fmt.Sprintf("INSERT INTO %s (user_id, title, description, content, created, status) VALUES (?, ?, ?, ?, ?, ?)",
		entity.BlogTableName)
	var created any
	if b.Created != nil {
		created = b.Created.T
	}
	res, err := m.DB.Exec(q, b.UserId, b.Title, b.Description, b.Content, created, b.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId() // IdType.AUTO writes the key back
	if err != nil {
		return err
	}
	b.Id = &id
	return nil
}

func (m *MySQLBlogMapper) UpdateById(b *entity.Blog) error {
	q := fmt.Sprintf("UPDATE %s SET user_id = ?, title = ?, description = ?, content = ?, created = ?, status = ? WHERE id = ?",
		entity.BlogTableName)
	var created any
	if b.Created != nil {
		created = b.Created.T
	}
	_, err := m.DB.Exec(q, b.UserId, b.Title, b.Description, b.Content, created, b.Status, b.Id)
	return err
}

const userColumns = "id, username, avatar, email, password, status, created, last_login"

// MySQLUserMapper is the production UserMapper.
type MySQLUserMapper struct{ DB *sql.DB }

func NewMySQLUserMapper(db *sql.DB) *MySQLUserMapper { return &MySQLUserMapper{DB: db} }

func scanUser(row interface{ Scan(...any) error }) (*entity.User, error) {
	var (
		u                             entity.User
		id                            sql.NullInt64
		username, avatar, email, pass sql.NullString
		status                        sql.NullInt32
		created, lastLogin            sql.NullTime
	)
	if err := row.Scan(&id, &username, &avatar, &email, &pass, &status, &created, &lastLogin); err != nil {
		return nil, err
	}
	if id.Valid {
		v := id.Int64
		u.Id = &v
	}
	if username.Valid {
		u.Username = username.String
	}
	if avatar.Valid {
		v := avatar.String
		u.Avatar = &v
	}
	if email.Valid {
		v := email.String
		u.Email = &v
	}
	if pass.Valid {
		v := pass.String
		u.Password = &v
	}
	if status.Valid {
		v := status.Int32
		u.Status = &v
	}
	if created.Valid {
		u.Created = lang.NewLocalDateTime(created.Time)
	}
	if lastLogin.Valid {
		u.LastLogin = lang.NewLocalDateTime(lastLogin.Time)
	}
	return &u, nil
}

func (m *MySQLUserMapper) SelectById(id int64) (*entity.User, error) {
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", userColumns, entity.UserTableName)
	u, err := scanUser(m.DB.QueryRow(q, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (m *MySQLUserMapper) SelectOne(qw *mybatisplus.QueryWrapper) (*entity.User, error) {
	where, args := qw.WhereSQL()
	q := fmt.Sprintf("SELECT %s FROM %s%s", userColumns, entity.UserTableName, where)
	u, err := scanUser(m.DB.QueryRow(q, args...))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

var _ = time.Time{}
