// Package exception mirrors com.markerhub.common.exception plus the Shiro and
// Spring exception types the handler dispatches on.
package exception

import "fmt"

// ShiroException mirrors org.apache.shiro.ShiroException - the root of every
// Shiro failure. GlobalExceptionHandler maps it to HTTP 401.
type ShiroException struct {
	Kind    string
	Message string
	Cause   error
}

func (e *ShiroException) Error() string { return e.Message }
func (e *ShiroException) Unwrap() error { return e.Cause }

// GetMessage mirrors Throwable.getMessage().
func (e *ShiroException) GetMessage() string { return e.Message }

// GetCause mirrors Throwable.getCause(); JwtFilter.onLoginFailure unwraps one
// level with `e.getCause() == null ? e : e.getCause()`.
func (e *ShiroException) GetCause() error { return e.Cause }

// The concrete Shiro subclasses this application can raise.
const (
	KindUnauthenticated    = "UnauthenticatedException"
	KindUnknownAccount     = "UnknownAccountException"
	KindLockedAccount      = "LockedAccountException"
	KindExpiredCredentials = "ExpiredCredentialsException"
	KindAuthentication     = "AuthenticationException"
)

// NewUnauthenticated reproduces the message Shiro's AuthenticatedAnnotationHandler
// throws for @RequiresAuthentication - note the two spaces after the period,
// which is verbatim from Shiro's source.
func NewUnauthenticated() *ShiroException {
	return &ShiroException{
		Kind:    KindUnauthenticated,
		Message: "The current Subject is not authenticated.  Access denied.",
	}
}

func NewUnknownAccount(msg string) *ShiroException {
	return &ShiroException{Kind: KindUnknownAccount, Message: msg}
}

func NewLockedAccount(msg string) *ShiroException {
	return &ShiroException{Kind: KindLockedAccount, Message: msg}
}

func NewExpiredCredentials(msg string) *ShiroException {
	return &ShiroException{Kind: KindExpiredCredentials, Message: msg}
}

// IllegalArgumentException mirrors java.lang.IllegalArgumentException, which
// org.springframework.util.Assert throws. Handled as HTTP 400.
type IllegalArgumentException struct{ Message string }

func (e *IllegalArgumentException) Error() string      { return e.Message }
func (e *IllegalArgumentException) GetMessage() string { return e.Message }

// nullPointerExceptionName is what Throwable.toString() yields for an NPE with
// no message, which is what Java 1.8 produces.
const nullPointerExceptionName = "java.lang.NullPointerException"

// NullPointerException mirrors java.lang.NullPointerException. The project
// targets Java 1.8 (pom.xml <java.version>1.8</java.version>), which predates
// helpful NPE messages, so getMessage() is null - and Result.fail(null) then
// serialises "msg":null. Message is therefore a *string, left nil.
type NullPointerException struct{ Message *string }

func (e *NullPointerException) Error() string {
	if e.Message == nil {
		return nullPointerExceptionName
	}
	return *e.Message
}

// RuntimeException mirrors any other java.lang.RuntimeException. Handled as 400.
type RuntimeException struct{ Message string }

func (e *RuntimeException) Error() string      { return e.Message }
func (e *RuntimeException) GetMessage() string { return e.Message }

// MethodArgumentNotValidException mirrors Spring's bean-validation failure.
// Errors preserves the order the validator reported them in; the handler takes
// findFirst().
type MethodArgumentNotValidException struct{ Errors []string }

func (e *MethodArgumentNotValidException) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Errors)
}
