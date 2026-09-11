package controller

import (
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/shiro"
)

// RequiresAuthentication mirrors Shiro's @RequiresAuthentication annotation,
// enforced by AuthorizationAttributeSourceAdvisor around the controller method.
//
// Shiro's AuthenticatedAnnotationHandler throws UnauthenticatedException when
// the subject is not authenticated; because the advisor runs inside
// DispatcherServlet, GlobalExceptionHandler catches it and returns 401.
func RequiresAuthentication(r *http.Request) {
	if !shiro.GetSubject(r.Context()).IsAuthenticated() {
		panic(exception.NewUnauthenticated())
	}
}
