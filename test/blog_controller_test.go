package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

type pageEnvelope struct {
	Records []struct {
		Id      int64  `json:"id"`
		UserId  int64  `json:"userId"`
		Title   string `json:"title"`
		Created string `json:"created"`
	} `json:"records"`
	Total       int64 `json:"total"`
	Size        int64 `json:"size"`
	Current     int64 `json:"current"`
	Pages       int64 `json:"pages"`
	SearchCount bool  `json:"searchCount"`
}

// TestBlogsIsPublic confirms /blogs needs no token: the jwt filter lets an
// anonymous request through and the handler carries no @RequiresAuthentication.
func TestBlogsIsPublic(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/blogs", "", "")

	assertResult(t, got, http.StatusOK, 200, "操作成功")

	var page pageEnvelope
	if err := json.Unmarshal(got.Data, &page); err != nil {
		t.Fatalf("data is not a Page: %v", err)
	}
	// new Page(currentPage, 5) over the six seeded rows.
	if page.Size != 5 || page.Current != 1 || page.Total != 6 || page.Pages != 2 {
		t.Errorf("page = size %d current %d total %d pages %d, want 5/1/6/2",
			page.Size, page.Current, page.Total, page.Pages)
	}
	if len(page.Records) != 5 {
		t.Fatalf("records = %d, want 5 (page size)", len(page.Records))
	}
	// orderByDesc("created"): 2020-10-20 is the newest seeded row.
	if page.Records[0].Id != 3 {
		t.Errorf("first record id = %d, want 3 (newest created)", page.Records[0].Id)
	}
	// @JsonFormat(pattern="yyyy-MM-dd") on Blog.created.
	if page.Records[0].Created != "2020-10-20" {
		t.Errorf("created = %q, want date-only %q", page.Records[0].Created, "2020-10-20")
	}
}

// TestBlogsPaging covers the currentPage parameter and its default.
func TestBlogsPaging(t *testing.T) {
	app := testsupport.New()

	second := do(t, app, http.MethodGet, "/blogs?currentPage=2", "", "")
	var page pageEnvelope
	if err := json.Unmarshal(second.Data, &page); err != nil {
		t.Fatalf("data is not a Page: %v", err)
	}
	if page.Current != 2 || len(page.Records) != 1 {
		t.Errorf("page 2 = current %d with %d records, want 2 with 1", page.Current, len(page.Records))
	}

	// Past the end: MP returns an empty record list, not an error.
	beyond := do(t, app, http.MethodGet, "/blogs?currentPage=99", "", "")
	var empty pageEnvelope
	if err := json.Unmarshal(beyond.Data, &empty); err != nil {
		t.Fatalf("data is not a Page: %v", err)
	}
	if len(empty.Records) != 0 {
		t.Errorf("records beyond the last page = %d, want 0", len(empty.Records))
	}
	if empty.Total != 6 {
		t.Errorf("total beyond the last page = %d, want 6", empty.Total)
	}
}

// TestBlogDetail covers the found and missing cases.
func TestBlogDetail(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/blog/1", "", "")
	assertResult(t, got, http.StatusOK, 200, "操作成功")

	var blog struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Created string `json:"created"`
	}
	if err := json.Unmarshal(got.Data, &blog); err != nil {
		t.Fatalf("data is not a Blog: %v", err)
	}
	if blog.Id != 1 {
		t.Errorf("id = %d, want 1", blog.Id)
	}
	if blog.Created != "2020-05-21" {
		t.Errorf("created = %q, want %q", blog.Created, "2020-05-21")
	}

	// Assert.notNull(blog, "该博客已被删除") -> IllegalArgumentException -> 400.
	missing := do(t, app, http.MethodGet, "/blog/9999", "", "")
	assertResult(t, missing, http.StatusBadRequest, 400, "该博客已被删除")
}

// TestBlogEditRequiresAuthentication guards the write endpoint.
func TestBlogEditRequiresAuthentication(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/blog/edit",
		`{"title":"t","description":"d","content":"c"}`, "")

	assertResult(t, got, http.StatusUnauthorized, 401,
		"The current Subject is not authenticated.  Access denied.")
}

// TestBlogEditCreate: with no id, a new row is inserted and userId/created/
// status are set server-side.
func TestBlogEditCreate(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/blog/edit",
		`{"title":"新标题","description":"新摘要","content":"新内容"}`, app.Token(1))

	assertResult(t, got, http.StatusOK, 200, "操作成功")
	if string(got.Data) != "null" {
		t.Errorf("data = %s, want null", got.Data)
	}

	list := do(t, app, http.MethodGet, "/blogs", "", "")
	var page pageEnvelope
	_ = json.Unmarshal(list.Data, &page)
	if page.Total != 7 {
		t.Errorf("total after create = %d, want 7", page.Total)
	}
}

// TestBlogEditUpdate: an existing id owned by the caller is updated, and the
// ignored properties are NOT taken from the request body.
func TestBlogEditUpdate(t *testing.T) {
	app := testsupport.New()

	// userId 999 and status 42 must be ignored by BeanUtil.copyProperties.
	got := do(t, app, http.MethodPost, "/blog/edit",
		`{"id":1,"userId":999,"status":42,"title":"改过的标题","description":"改过的摘要","content":"改过的内容"}`,
		app.Token(1))
	assertResult(t, got, http.StatusOK, 200, "操作成功")

	detail := do(t, app, http.MethodGet, "/blog/1", "", "")
	var blog struct {
		Id     int64  `json:"id"`
		UserId int64  `json:"userId"`
		Status int32  `json:"status"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(detail.Data, &blog); err != nil {
		t.Fatalf("data is not a Blog: %v", err)
	}
	if blog.Title != "改过的标题" {
		t.Errorf("title = %q, want the updated value", blog.Title)
	}
	if blog.UserId != 1 {
		t.Errorf("userId = %d, want 1 - copyProperties must ignore userId", blog.UserId)
	}
	if blog.Status != 0 {
		t.Errorf("status = %d, want 0 - copyProperties must ignore status", blog.Status)
	}
}

// TestBlogEditForeignBlog: Assert.isTrue(temp.userId == profile.id, "没有权限编辑").
func TestBlogEditForeignBlog(t *testing.T) {
	app := testsupport.New()

	// Token for user 2, editing blog 1 which belongs to user 1.
	app.UserMapper.Seed(seedUser(2, "someone"))

	got := do(t, app, http.MethodPost, "/blog/edit",
		`{"id":1,"title":"t","description":"d","content":"c"}`, app.Token(2))

	assertResult(t, got, http.StatusBadRequest, 400, "没有权限编辑")
}

// TestBlogEditMissingId reproduces a genuine defect in the Java source: editing
// a non-existent id dereferences a null temp, throwing NullPointerException.
// Under Java 8 that exception carries no message, so Result.fail(e.getMessage())
// yields "msg":null with HTTP 400. Preserved rather than "fixed" so behaviour
// matches the original exactly.
func TestBlogEditMissingIdThrowsNPEWithNullMessage(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/blog/edit",
		`{"id":9999,"title":"t","description":"d","content":"c"}`, app.Token(1))

	if got.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", got.Status)
	}
	if got.Code != 400 {
		t.Errorf("code = %d, want 400", got.Code)
	}
	if got.Msg != nil {
		t.Errorf("msg = %q, want JSON null (Java 8 NPE has no message)", *got.Msg)
	}
}

// TestBlogEditValidation covers @NotBlank on Blog.
func TestBlogEditValidation(t *testing.T) {
	cases := []struct{ name, body, wantMsg string }{
		{"blank title", `{"title":"","description":"d","content":"c"}`, "标题不能为空"},
		{"blank description", `{"title":"t","description":"","content":"c"}`, "摘要不能为空"},
		{"blank content", `{"title":"t","description":"d","content":""}`, "内容不能为空"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := testsupport.New()
			got := do(t, app, http.MethodPost, "/blog/edit", tc.body, app.Token(1))
			assertResult(t, got, http.StatusBadRequest, 400, tc.wantMsg)
		})
	}
}
