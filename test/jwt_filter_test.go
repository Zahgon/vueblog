package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/markerhub/vueblog-go/internal/testsupport"
	"github.com/markerhub/vueblog-go/internal/util"
)

// TestExpiredTokenRejected: JwtFilter throws ExpiredCredentialsException, a
// ShiroException, so the handler returns 401 with that message.
func TestExpiredTokenRejected(t *testing.T) {
	app := testsupport.New()

	expiring := &util.JwtUtils{Secret: app.JwtUtils.Secret, Expire: -3600, Header: "Authorization"}
	token, err := expiring.GenerateToken(1)
	if err != nil {
		t.Fatalf("failed to mint an expired token: %v", err)
	}

	got := do(t, app, http.MethodGet, "/user/index", "", token)

	assertResult(t, got, http.StatusUnauthorized, 401, "token已失效，请重新登录")
}

// TestMalformedTokenRejected: getClaimByToken returns null, which the filter
// turns into the same ExpiredCredentialsException.
func TestMalformedTokenRejected(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/user/index", "", "not.a.jwt")

	assertResult(t, got, http.StatusUnauthorized, 401, "token已失效，请重新登录")
}

// TestTokenSignedWithWrongSecretRejected guards against signature bypass.
func TestTokenSignedWithWrongSecretRejected(t *testing.T) {
	app := testsupport.New()

	attacker := &util.JwtUtils{Secret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", Expire: 604800}
	token, err := attacker.GenerateToken(1)
	if err != nil {
		t.Fatalf("failed to mint a foreign token: %v", err)
	}

	got := do(t, app, http.MethodGet, "/user/index", "", token)

	assertResult(t, got, http.StatusUnauthorized, 401, "token已失效，请重新登录")
}

// TestValidTokenForMissingUser: the realm throws UnknownAccountException, which
// onLoginFailure writes directly. Faithful quirk - that path never sets a status
// code, so the response is HTTP 200 carrying code 400.
func TestValidTokenForMissingUserWritesBodyWith200(t *testing.T) {
	app := testsupport.New()

	got := do(t, app, http.MethodGet, "/user/index", "", app.Token(4242))

	assertResult(t, got, http.StatusOK, 400, "账户不存在")
}

// TestLockedAccount: status == -1 throws LockedAccountException on the same path.
func TestLockedAccountWritesBodyWith200(t *testing.T) {
	app := testsupport.New()

	locked := seedUser(3, "locked")
	lockedStatus := int32(-1)
	locked.Status = &lockedStatus
	app.UserMapper.Seed(locked)

	got := do(t, app, http.MethodGet, "/user/index", "", app.Token(3))

	assertResult(t, got, http.StatusOK, 400, "账户已被锁定")
}

// TestOptionsShortCircuits: preHandle answers CORS preflight with 200 and an
// empty body before authentication runs.
func TestOptionsShortCircuits(t *testing.T) {
	app := testsupport.New()

	req := httptest.NewRequest(http.MethodOptions, "/user/index", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	rec := httptest.NewRecorder()

	app.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("OPTIONS status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("OPTIONS body = %q, want empty", rec.Body.String())
	}
	if o := rec.Header().Get("Access-control-Allow-Origin"); o != "http://localhost:8080" {
		t.Errorf("Access-control-Allow-Origin = %q, want the echoed Origin", o)
	}
	if m := rec.Header().Get("Access-Control-Allow-Methods"); m != "GET,POST,OPTIONS,PUT,DELETE" {
		t.Errorf("Access-Control-Allow-Methods = %q, want the JwtFilter list", m)
	}
}

// TestEndToEndLoginThenAuthenticatedCall exercises the full round trip: the JWT
// minted by /login must be accepted by the filter on a protected endpoint.
func TestEndToEndLoginThenAuthenticatedCall(t *testing.T) {
	app := testsupport.New()

	login := do(t, app, http.MethodPost, "/login",
		`{"username":"markerhub","password":"`+testsupport.SeedPasswordPlain+`"}`, "")
	token := login.Header.Get("Authorization")
	if token == "" {
		t.Fatal("login did not return a token")
	}

	got := do(t, app, http.MethodGet, "/user/index", "", token)
	assertResult(t, got, http.StatusOK, 200, "操作成功")
}

// TestLogoutDoesNotInvalidateTheToken pins genuinely surprising behaviour that
// both implementations share: authentication is stateless, so the JWT keeps
// working after /logout. Shiro's session teardown does not revoke it.
func TestLogoutDoesNotInvalidateTheToken(t *testing.T) {
	app := testsupport.New()
	token := app.Token(1)

	if got := do(t, app, http.MethodGet, "/logout", "", token); got.Code != 200 {
		t.Fatalf("logout failed: %s", got.Raw)
	}

	got := do(t, app, http.MethodGet, "/user/index", "", token)
	assertResult(t, got, http.StatusOK, 200, "操作成功")
}

var _ = time.Now
