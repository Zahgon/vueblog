package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

// TestLoginSuccess covers AccountController.login on the happy path: it must
// return the four-key profile map and set both response headers.
func TestLoginSuccess(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/login",
		`{"username":"markerhub","password":"`+testsupport.SeedPasswordPlain+`"}`, "")

	assertResult(t, got, http.StatusOK, 200, "操作成功")

	var data struct {
		Id       int64   `json:"id"`
		Username string  `json:"username"`
		Avatar   *string `json:"avatar"`
		Email    *string `json:"email"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data is not the profile map: %v", err)
	}
	if data.Id != 1 || data.Username != "markerhub" {
		t.Errorf("profile = %+v, want id 1 / markerhub", data)
	}
	// m_user.email is NULL in the seed data; Jackson's ALWAYS inclusion emits it.
	if data.Email != nil {
		t.Errorf("email = %v, want null", *data.Email)
	}

	// response.setHeader("Authorization", jwt)
	if got.Header.Get("Authorization") == "" {
		t.Error("Authorization header not set on login response")
	}
	if h := got.Header.Get("Access-control-Expose-Headers"); h != "Authorization" {
		t.Errorf("Access-control-Expose-Headers = %q, want %q", h, "Authorization")
	}
}

// TestLoginWrongPassword: the Java code RETURNS Result.fail rather than
// throwing, so the HTTP status stays 200 while the body carries code 400.
func TestLoginWrongPassword(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/login",
		`{"username":"markerhub","password":"wrong"}`, "")

	assertResult(t, got, http.StatusOK, 400, "密码不正确")

	if got.Header.Get("Authorization") != "" {
		t.Error("Authorization header must not be set when the password is wrong")
	}
}

// TestLoginUnknownUser: Assert.notNull throws IllegalArgumentException -> 400.
func TestLoginUnknownUser(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodPost, "/login",
		`{"username":"nobody","password":"111111"}`, "")

	assertResult(t, got, http.StatusBadRequest, 400, "用户不存在")
}

// TestLoginValidation covers @NotBlank on LoginDto. The handler reports only
// the first error, and field declaration order decides which that is.
func TestLoginValidation(t *testing.T) {
	cases := []struct {
		name, body, wantMsg string
	}{
		{"blank username", `{"username":"","password":"111111"}`, "昵称不能为空"},
		{"whitespace username", `{"username":"   ","password":"111111"}`, "昵称不能为空"},
		{"missing username", `{"password":"111111"}`, "昵称不能为空"},
		{"blank password", `{"username":"markerhub","password":""}`, "密码不能为空"},
		{"both blank reports username first", `{"username":"","password":""}`, "昵称不能为空"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := testsupport.New()
			got := do(t, app, http.MethodPost, "/login", tc.body, "")
			assertResult(t, got, http.StatusBadRequest, 400, tc.wantMsg)
		})
	}
}

// TestLogoutRequiresAuthentication: @RequiresAuthentication throws
// UnauthenticatedException, which the ShiroException handler maps to 401 with
// Shiro's own message.
func TestLogoutRequiresAuthentication(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/logout", "", "")

	assertResult(t, got, http.StatusUnauthorized, 401,
		"The current Subject is not authenticated.  Access denied.")
}

// TestLogoutAuthenticated returns Result.succ(null) -> "data":null.
func TestLogoutAuthenticated(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/logout", "", app.Token(1))

	assertResult(t, got, http.StatusOK, 200, "操作成功")
	if string(got.Data) != "null" {
		t.Errorf("data = %s, want null", got.Data)
	}
}
