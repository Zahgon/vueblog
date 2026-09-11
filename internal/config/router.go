package config

import (
	"encoding/json"
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/controller"
	"github.com/markerhub/vueblog-go/internal/shiro"
)

// Handlers bundles the three @RestController beans and the Shiro filter.
type Handlers struct {
	Account *controller.AccountController
	Blog    *controller.BlogController
	User    *controller.UserController
	Filter  *shiro.JwtFilter
}

// handlerFunc is the shape of a Spring @RestController method: it returns a
// Result which the framework serialises.
type handlerFunc func(w http.ResponseWriter, r *http.Request) *lang.Result

// NewRouter reproduces the servlet stack in the order Spring Boot runs it:
//
//	shiroFilter ("/**" -> jwt)      <- a servlet Filter, runs first
//	  DispatcherServlet
//	    CORS from CorsConfig        <- WebMvcConfigurer, inside the dispatcher
//	      @RestControllerAdvice     <- GlobalExceptionHandler
//	        controller method
//
// The order matters: because the jwt filter sits outside DispatcherServlet and
// answers OPTIONS itself, a preflight response carries the filter's CORS header
// values, not CorsConfig's. Requests that reach a handler get CorsConfig's
// values layered on top.
func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /login", h.dispatch(h.Account.Login))
	mux.Handle("GET /logout", h.dispatch(h.Account.Logout))
	mux.Handle("GET /blogs", h.dispatch(h.Blog.List))
	mux.Handle("GET /blog/{id}", h.dispatch(func(w http.ResponseWriter, r *http.Request) *lang.Result {
		return h.Blog.Detail(w, r, r.PathValue("id"))
	}))
	mux.Handle("POST /blog/edit", h.dispatch(h.Blog.Edit))
	mux.Handle("GET /user/index", h.dispatch(h.User.Index))
	mux.Handle("POST /user/save", h.dispatch(h.User.Save))
	mux.Handle("/", h.notFound())

	// The Shiro filter chain wraps everything, matching "/**".
	return h.shiroFilterChain(mux)
}

// shiroFilterChain is the servlet Filter layer: it runs before DispatcherServlet
// for every path and binds the authenticated Subject onto the request.
func (h *Handlers) shiroFilterChain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GlobalExceptionHandler is an @RestControllerAdvice and therefore only
		// covers exceptions raised inside DispatcherServlet. This recover sits
		// outside it so that ExpiredCredentialsException - thrown by the filter
		// itself - is still rendered as the 401 the application intends.
		defer func() {
			if rec := recover(); rec != nil {
				exception.Handle(w, rec)
			}
		}()

		if !h.Filter.PreHandle(w, r) {
			return // OPTIONS answered with 200 before authentication
		}

		ok, subject := h.Filter.OnAccessDenied(w, r)
		if !ok {
			return // onLoginFailure already wrote the body
		}

		next.ServeHTTP(w, shiro.WithSubject(r, subject))
	})
}

// dispatch is the DispatcherServlet layer for one handler method.
func (h *Handlers) dispatch(fn handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				exception.Handle(w, rec)
			}
		}()

		Cors(w, r) // CorsConfig, applied once a handler has been matched

		result := fn(w, r)
		writeJSON(w, http.StatusOK, result)
	})
}

// notFound mirrors Spring Boot's BasicErrorController for an unmapped path.
func (h *Handlers) notFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cors(w, r)
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": 404,
			"error":  "Not Found",
			"path":   r.URL.Path,
		})
	})
}

// writeJSON serialises a controller return value the way Spring's
// MappingJackson2HttpMessageConverter does.
func writeJSON(w http.ResponseWriter, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
