package shiro

import (
	"encoding/json"
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/util"
)

// JwtFilter mirrors com.markerhub.shiro.JwtFilter extends AuthenticatingFilter.
type JwtFilter struct {
	JwtUtils *util.JwtUtils
	Realm    *AccountRealm
}

func NewJwtFilter(j *util.JwtUtils, realm *AccountRealm) *JwtFilter {
	return &JwtFilter{JwtUtils: j, Realm: realm}
}

// CreateToken mirrors JwtFilter.createToken: null when the header is absent.
// StringUtils.isEmpty treats null and "" as empty but NOT "   ".
func (f *JwtFilter) CreateToken(r *http.Request) *JwtToken {
	jwtHeader := r.Header.Get("Authorization")
	if jwtHeader == "" {
		return nil
	}
	return NewJwtToken(jwtHeader)
}

// PreHandle mirrors JwtFilter.preHandle: it echoes the CORS headers and
// short-circuits OPTIONS with 200 before any authentication runs.
// Returning false means "stop the chain".
func (f *JwtFilter) PreHandle(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return false
	}
	return true
}

// OnAccessDenied mirrors JwtFilter.onAccessDenied.
//
// Returning true lets the request continue - note that an absent token does NOT
// reject the request here. Anonymous access proceeds and is only stopped later
// by @RequiresAuthentication on the individual handler. That is what makes
// /blogs and /blog/{id} publicly readable.
func (f *JwtFilter) OnAccessDenied(w http.ResponseWriter, r *http.Request) (bool, *Subject) {
	jwtHeader := r.Header.Get("Authorization")
	if jwtHeader == "" {
		return true, &Subject{}
	}

	claims := f.JwtUtils.GetClaimByToken(jwtHeader)
	if claims == nil || f.JwtUtils.IsTokenExpired(claims.Expiration) {
		panic(exception.NewExpiredCredentials("token已失效，请重新登录"))
	}

	return f.executeLogin(w, r, jwtHeader)
}

// executeLogin mirrors AuthenticatingFilter.executeLogin: build the token, run
// the realm, and on failure hand off to onLoginFailure.
func (f *JwtFilter) executeLogin(w http.ResponseWriter, r *http.Request, jwtHeader string) (ok bool, subject *Subject) {
	token := NewJwtToken(jwtHeader)

	defer func() {
		if rec := recover(); rec != nil {
			shiroErr, isShiro := rec.(*exception.ShiroException)
			if !isShiro {
				panic(rec) // NPE / IllegalArgument propagate to the global handler
			}
			f.OnLoginFailure(w, shiroErr)
			ok, subject = false, nil
		}
	}()

	profile := f.Realm.DoGetAuthenticationInfo(token)
	return true, &Subject{principal: profile, authenticated: true}
}

// OnLoginFailure mirrors JwtFilter.onLoginFailure: it writes the JSON body
// directly to the response and returns false.
//
// Faithful detail: it never sets a status code, so the response is HTTP 200
// carrying a body with "code":400. The Java code writes via
// httpServletResponse.getWriter().print(json) with no setStatus call, and
// Shiro's filter has not committed any other status by that point.
func (f *JwtFilter) OnLoginFailure(w http.ResponseWriter, e *exception.ShiroException) {
	// `e.getCause() == null ? e : e.getCause()` - unwrap exactly one level.
	message := e.GetMessage()
	if e.GetCause() != nil {
		message = e.GetCause().Error()
	}

	result := lang.Fail(message)
	body, err := json.Marshal(result)
	if err != nil {
		return // the Java catch block swallows IOException silently
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	_, _ = w.Write(body)
}
