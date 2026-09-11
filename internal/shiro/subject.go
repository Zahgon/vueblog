// Package shiro mirrors com.markerhub.shiro plus the small slice of Apache
// Shiro's runtime that vueblog depends on (Subject, realm, authenticating
// filter). Shiro binds the Subject to a ThreadLocal; Go's equivalent is a
// request-scoped context value, so the Subject travels on the request context.
package shiro

import (
	"context"
	"net/http"
)

// AccountProfile mirrors com.markerhub.shiro.AccountProfile.
type AccountProfile struct {
	Id       *int64  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Email    *string `json:"email"`
}

// GetId mirrors AccountProfile.getId().
func (p *AccountProfile) GetId() *int64 { return p.Id }

// JwtToken mirrors com.markerhub.shiro.JwtToken: an AuthenticationToken whose
// principal and credentials are both the raw JWT string.
type JwtToken struct{ token string }

func NewJwtToken(jwt string) *JwtToken { return &JwtToken{token: jwt} }

func (t *JwtToken) GetPrincipal() any   { return t.token }
func (t *JwtToken) GetCredentials() any { return t.token }

// Subject mirrors org.apache.shiro.subject.Subject for the operations vueblog
// uses: isAuthenticated, getPrincipal and logout.
type Subject struct {
	principal     *AccountProfile
	authenticated bool
}

func (s *Subject) IsAuthenticated() bool { return s != nil && s.authenticated }
func (s *Subject) GetPrincipal() *AccountProfile {
	if s == nil {
		return nil
	}
	return s.principal
}

// Logout mirrors Subject.logout(): clears the principal and authenticated flag.
// Shiro also drops the Redis-backed session; SessionStore handles that side.
func (s *Subject) Logout() {
	if s == nil {
		return
	}
	s.principal = nil
	s.authenticated = false
}

type subjectKeyType struct{}

var subjectKey subjectKeyType

// WithSubject binds a Subject to the request context - the analogue of Shiro
// binding it to the current thread.
func WithSubject(r *http.Request, s *Subject) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), subjectKey, s))
}

// GetSubject mirrors SecurityUtils.getSubject(). Shiro always returns a Subject
// (anonymous when unauthenticated), never null, so this does too.
func GetSubject(ctx context.Context) *Subject {
	if s, ok := ctx.Value(subjectKey).(*Subject); ok && s != nil {
		return s
	}
	return &Subject{}
}
