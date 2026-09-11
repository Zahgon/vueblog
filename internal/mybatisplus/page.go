// Package mybatisplus reproduces the parts of com.baomidou.mybatisplus that
// vueblog actually uses: Page (which is serialised straight into the HTTP
// response by BlogController.list) and QueryWrapper.
package mybatisplus

// Page mirrors com.baomidou.mybatisplus.extension.plugins.pagination.Page
// at version 3.2.0, the version pinned in pom.xml.
//
// BlogController.list returns the Page itself inside Result.data, so this
// struct's JSON shape is part of the public API contract. The field set and
// order below follow MP 3.2.0's getters: records, total, size, current, ascs,
// descs, optimizeCountSql, searchCount, then the derived pages.
type Page[T any] struct {
	Records          []T      `json:"records"`
	Total            int64    `json:"total"`
	Size             int64    `json:"size"`
	Current          int64    `json:"current"`
	Ascs             []string `json:"ascs"`
	Descs            []string `json:"descs"`
	OptimizeCountSql bool     `json:"optimizeCountSql"`
	SearchCount      bool     `json:"searchCount"`
	Pages            int64    `json:"pages"`
}

// NewPage mirrors `new Page(currentPage, 5)`. MP defaults optimizeCountSql and
// isSearchCount to true and leaves records as an empty list.
func NewPage[T any](current, size int64) *Page[T] {
	return &Page[T]{
		Records:          []T{},
		Size:             size,
		Current:          current,
		OptimizeCountSql: true,
		SearchCount:      true,
	}
}

// ComputePages mirrors IPage.getPages(): ceil(total/size), or 0 when size <= 0.
// MP computes this in the getter, so it must be refreshed after Total is set.
func (p *Page[T]) ComputePages() {
	if p.Size == 0 {
		p.Pages = 0
		return
	}
	pages := p.Total / p.Size
	if p.Total%p.Size != 0 {
		pages++
	}
	p.Pages = pages
}

// Offset mirrors the LIMIT offset MP's PaginationInterceptor computes.
// MP clamps current to a minimum of 1 before multiplying.
func (p *Page[T]) Offset() int64 {
	if p.Current <= 1 {
		return 0
	}
	return (p.Current - 1) * p.Size
}
