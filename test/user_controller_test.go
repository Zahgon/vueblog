package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

// TestUserIndexRequiresAuthentication guards /user/index.
func TestUserIndexRequiresAuthentication(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/user/index", "", "")

	assertResult(t, got, http.StatusUnauthorized, 401,
		"The current Subject is not authenticated.  Access denied.")
}

// TestUserIndexReturnsUserOne pins a quirk of the original: the handler is
// hardcoded to getById(1L) and ignores the authenticated principal entirely.
func TestUserIndexReturnsUserOne(t *testing.T) {
	app := testsupport.New()
	app.UserMapper.Seed(seedUser(2, "someone"))

	// Authenticated as user 2, yet user 1 comes back.
	got := do(t, app, http.MethodGet, "/user/index", "", app.Token(2))
	assertResult(t, got, http.StatusOK, 200, "操作成功")

	var user struct {
		Id       int64   `json:"id"`
		Username string  `json:"username"`
		Created  string  `json:"created"`
		Email    *string `json:"email"`
	}
	if err := json.Unmarshal(got.Data, &user); err != nil {
		t.Fatalf("data is not a User: %v", err)
	}
	if user.Id != 1 || user.Username != "markerhub" {
		t.Errorf("user = id %d / %s, want 1 / markerhub (endpoint is hardcoded to id 1)",
			user.Id, user.Username)
	}
	// An unannotated LocalDateTime serialises as ISO-8601 without an offset.
	if user.Created != "2020-04-20T10:44:01" {
		t.Errorf("created = %q, want %q", user.Created, "2020-04-20T10:44:01")
	}
	if user.Email != nil {
		t.Errorf("email = %q, want null", *user.Email)
	}
}

// TestUserSaveIsPublic: /user/save has no @RequiresAuthentication - it
// validates and echoes the payload back without persisting it.
func TestUserSaveIsPublic(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/user/save",
		`{"username":"newbie","email":"a@b.com"}`, "")

	assertResult(t, got, http.StatusOK, 200, "操作成功")

	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.Unmarshal(got.Data, &user); err != nil {
		t.Fatalf("data is not a User: %v", err)
	}
	if user.Username != "newbie" || user.Email != "a@b.com" {
		t.Errorf("echoed user = %+v, want the submitted values", user)
	}

	// Nothing was persisted: /login for that username still fails.
	login := do(t, app, http.MethodPost, "/login", `{"username":"newbie","password":"x"}`, "")
	assertResult(t, login, http.StatusBadRequest, 400, "用户不存在")
}

// TestUserSaveValidation covers @NotBlank and @Email on User.
func TestUserSaveValidation(t *testing.T) {
	cases := []struct{ name, body, wantMsg string }{
		{"blank username", `{"username":"","email":"a@b.com"}`, "昵称不能为空"},
		{"missing email", `{"username":"newbie"}`, "邮箱不能为空"},
		{"blank email", `{"username":"newbie","email":""}`, "邮箱不能为空"},
		{"malformed email", `{"username":"newbie","email":"not-an-email"}`, "邮箱格式不正确"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := testsupport.New()
			got := do(t, app, http.MethodPost, "/user/save", tc.body, "")
			assertResult(t, got, http.StatusBadRequest, 400, tc.wantMsg)
		})
	}
}
