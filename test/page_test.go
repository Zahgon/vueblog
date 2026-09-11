package test

import (
	"testing"

	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// TestPageDefaults mirrors `new Page(current, size)`: MP defaults both
// optimizeCountSql and isSearchCount to true.
func TestPageDefaults(t *testing.T) {
	p := mybatisplus.NewPage[*entity.Blog](1, 5)

	if p.Current != 1 || p.Size != 5 {
		t.Errorf("page = current %d size %d, want 1/5", p.Current, p.Size)
	}
	if !p.OptimizeCountSql || !p.SearchCount {
		t.Error("optimizeCountSql and searchCount must both default to true")
	}
	if p.Records == nil {
		t.Error("records must default to an empty list, not null")
	}
}

// TestComputePages mirrors IPage.getPages(): ceil(total/size).
func TestComputePages(t *testing.T) {
	cases := []struct{ total, size, want int64 }{
		{0, 5, 0},
		{1, 5, 1},
		{5, 5, 1},
		{6, 5, 2},
		{10, 5, 2},
		{11, 5, 3},
		{7, 0, 0}, // guard against divide-by-zero
	}
	for _, tc := range cases {
		p := mybatisplus.NewPage[*entity.Blog](1, tc.size)
		p.Total = tc.total
		p.ComputePages()
		if p.Pages != tc.want {
			t.Errorf("pages(total=%d,size=%d) = %d, want %d", tc.total, tc.size, p.Pages, tc.want)
		}
	}
}

// TestPageOffset mirrors the LIMIT offset PaginationInterceptor computes,
// including MP's clamp of current to a minimum of 1.
func TestPageOffset(t *testing.T) {
	cases := []struct{ current, size, want int64 }{
		{1, 5, 0},
		{2, 5, 5},
		{3, 5, 10},
		{0, 5, 0},  // clamped
		{-1, 5, 0}, // clamped
	}
	for _, tc := range cases {
		p := mybatisplus.NewPage[*entity.Blog](tc.current, tc.size)
		if got := p.Offset(); got != tc.want {
			t.Errorf("offset(current=%d,size=%d) = %d, want %d", tc.current, tc.size, got, tc.want)
		}
	}
}

// TestQueryWrapperSQL covers the two builder methods vueblog uses.
func TestQueryWrapperSQL(t *testing.T) {
	qw := mybatisplus.NewQueryWrapper().Eq("username", "markerhub")
	where, args := qw.WhereSQL()
	if where != " WHERE username = ?" {
		t.Errorf("where = %q", where)
	}
	if len(args) != 1 || args[0] != "markerhub" {
		t.Errorf("args = %v, want [markerhub]", args)
	}

	ordered := mybatisplus.NewQueryWrapper().OrderByDesc("created")
	if got := ordered.OrderSQL(); got != " ORDER BY created DESC" {
		t.Errorf("order = %q", got)
	}

	empty := mybatisplus.NewQueryWrapper()
	if w, a := empty.WhereSQL(); w != "" || a != nil {
		t.Errorf("empty wrapper produced %q / %v", w, a)
	}
}
