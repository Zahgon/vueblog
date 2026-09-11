package test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/shiro"
)

// The Java originals of these are exercised by the framework rather than by
// application code: Throwable accessors, Jackson's deserialisation side, the
// entity setters MyBatis-Plus calls, and the Realm hooks Shiro invokes. Nothing
// in the port's own suite reached them.

func TestShiroExceptionUnwrapAndGetMessage(t *testing.T) {
	cause := errors.New("token expired")
	e := &exception.ShiroException{
		Kind:    exception.KindExpiredCredentials,
		Message: "credentials expired",
		Cause:   cause,
	}

	if got := e.Error(); got != "credentials expired" {
		t.Errorf("Error() = %q, want %q", got, "credentials expired")
	}
	if got := e.GetMessage(); got != "credentials expired" {
		t.Errorf("GetMessage() = %q, want %q", got, "credentials expired")
	}
	if got := e.Unwrap(); got != cause {
		t.Errorf("Unwrap() = %v, want %v", got, cause)
	}
	// Unwrap is what makes errors.Is work across the wrapper, which is the
	// Go equivalent of walking getCause().
	if !errors.Is(e, cause) {
		t.Error("errors.Is(e, cause) = false, want true")
	}
}

func TestShiroExceptionUnwrapWithoutCause(t *testing.T) {
	e := exception.NewUnauthenticated()
	if e.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", e.Unwrap())
	}
	// Verbatim from Shiro's source, including the two spaces after the period.
	want := "The current Subject is not authenticated.  Access denied."
	if e.GetMessage() != want {
		t.Errorf("GetMessage() = %q, want %q", e.GetMessage(), want)
	}
}

func TestLocalDateTimeUnmarshalJSONAndString(t *testing.T) {
	var l lang.LocalDateTime
	if err := json.Unmarshal([]byte(`"2020-04-12T15:04:05"`), &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := l.String(); got != "2020-04-12T15:04:05" {
		t.Errorf("String() = %q, want %q", got, "2020-04-12T15:04:05")
	}
	// Round-trips back to the same bytes Jackson would write.
	out, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `"2020-04-12T15:04:05"` {
		t.Errorf("marshal = %s, want %q", out, `"2020-04-12T15:04:05"`)
	}
}

func TestLocalDateTimeUnmarshalJSONNullAndEmpty(t *testing.T) {
	for _, raw := range []string{`"null"`, `""`} {
		var l lang.LocalDateTime
		if err := json.Unmarshal([]byte(raw), &l); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		if !l.T.IsZero() {
			t.Errorf("unmarshal %s left %v, want zero time", raw, l.T)
		}
	}
}

func TestLocalDateTimeUnmarshalJSONRejectsBadValue(t *testing.T) {
	var l lang.LocalDateTime
	if err := json.Unmarshal([]byte(`"12/04/2020"`), &l); err == nil {
		t.Error("unmarshal of a non-ISO value succeeded, want error")
	}
}

func TestLocalDateUnmarshalJSONAndString(t *testing.T) {
	var d lang.LocalDate
	if err := json.Unmarshal([]byte(`"2020-04-12"`), &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// LocalDate keeps full precision in memory - Blog.created is compared and
	// stored as a datetime - and only *serialises* date-only under
	// @JsonFormat(pattern="yyyy-MM-dd"). String() therefore shows the time.
	if got := d.String(); got != "2020-04-12T00:00:00" {
		t.Errorf("String() = %q, want %q", got, "2020-04-12T00:00:00")
	}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `"2020-04-12"` {
		t.Errorf("marshal = %s, want %q", out, `"2020-04-12"`)
	}
}

func TestBlogSetId(t *testing.T) {
	// MyBatis-Plus setters return the entity for chaining; SetId also has to
	// take the address of its argument because Blog.Id is a pointer.
	b := (&entity.Blog{}).SetId(7).SetTitle("hello")
	if b.Id == nil || *b.Id != 7 {
		t.Errorf("Id = %v, want 7", b.Id)
	}
	if b.Title != "hello" {
		t.Errorf("Title = %q, want %q", b.Title, "hello")
	}
}

func TestAccountRealmSupports(t *testing.T) {
	r := &shiro.AccountRealm{}
	if !r.Supports(&shiro.JwtToken{}) {
		t.Error("Supports(*JwtToken) = false, want true")
	}
	if r.Supports("not a token") {
		t.Error("Supports(string) = true, want false")
	}
	if r.Supports(nil) {
		t.Error("Supports(nil) = true, want false")
	}
}

func TestAccountRealmDoGetAuthorizationInfo(t *testing.T) {
	// The Java override returns null: vueblog does no role/permission checking.
	r := &shiro.AccountRealm{}
	if got := r.DoGetAuthorizationInfo(); got != nil {
		t.Errorf("DoGetAuthorizationInfo() = %v, want nil", got)
	}
}

func TestLocalDateTimeMarshalUsesSecondPrecision(t *testing.T) {
	l := lang.NewLocalDateTime(time.Date(2020, 4, 12, 15, 4, 5, 123456789, time.Local))
	out, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// The MySQL DATETIME columns are second-precision, so nanoseconds are dropped.
	if string(out) != `"2020-04-12T15:04:05"` {
		t.Errorf("marshal = %s, want %q", out, `"2020-04-12T15:04:05"`)
	}
}

func TestIllegalArgumentAndRuntimeExceptionAccessors(t *testing.T) {
	illegal := &exception.IllegalArgumentException{Message: "用户不存在"}
	if illegal.Error() != "用户不存在" || illegal.GetMessage() != "用户不存在" {
		t.Errorf("IllegalArgumentException = %q/%q", illegal.Error(), illegal.GetMessage())
	}

	runtime := &exception.RuntimeException{Message: "token已失效，请重新登录"}
	if runtime.Error() != "token已失效，请重新登录" || runtime.GetMessage() != "token已失效，请重新登录" {
		t.Errorf("RuntimeException = %q/%q", runtime.Error(), runtime.GetMessage())
	}
}

func TestNullPointerExceptionMessageIsNilOnJava8(t *testing.T) {
	// Java 1.8 predates helpful NPE messages, so getMessage() is null and
	// Result.fail(null) serialises "msg":null.
	npe := &exception.NullPointerException{}
	if got := npe.Error(); got != "java.lang.NullPointerException" {
		t.Errorf("Error() = %q", got)
	}

	message := "explicit"
	explicit := &exception.NullPointerException{Message: &message}
	if got := explicit.Error(); got != "explicit" {
		t.Errorf("Error() = %q, want %q", got, "explicit")
	}
}

func TestMethodArgumentNotValidExceptionKeepsValidatorOrder(t *testing.T) {
	err := &exception.MethodArgumentNotValidException{Errors: []string{"标题不能为空", "内容不能为空"}}
	want := "validation failed: [标题不能为空 内容不能为空]"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestJwtFilterCreateToken(t *testing.T) {
	filter := &shiro.JwtFilter{}

	// No Authorization header: JwtFilter.createToken returns null.
	bare := httptest.NewRequest(http.MethodGet, "/blogs", nil)
	if token := filter.CreateToken(bare); token != nil {
		t.Errorf("CreateToken with no header = %v, want nil", token)
	}

	withHeader := httptest.NewRequest(http.MethodGet, "/blogs", nil)
	withHeader.Header.Set("Authorization", "a.b.c")
	token := filter.CreateToken(withHeader)
	if token == nil {
		t.Fatal("CreateToken with a header = nil, want a token")
	}
	if got := token.GetCredentials(); got != "a.b.c" {
		t.Errorf("GetCredentials() = %v, want %q", got, "a.b.c")
	}
}
