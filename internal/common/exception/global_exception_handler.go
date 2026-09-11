package exception

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/lang"
)

// Handle mirrors com.markerhub.common.exception.GlobalExceptionHandler.
//
// Spring dispatches to the @ExceptionHandler whose declared type most closely
// matches the thrown exception, so the order below (most specific first) is
// what reproduces Spring's selection:
//
//	ShiroException                  -> 401
//	MethodArgumentNotValidException -> 400, first field error message
//	IllegalArgumentException        -> 400
//	RuntimeException                -> 400  (catches NullPointerException too)
//
// The @ResponseStatus annotation sets the HTTP status while the body carries
// its own code, which is why 401 responses have code 401 in the body and 400
// responses carry Result.fail's default code of 400.
func Handle(w http.ResponseWriter, rec any) {
	status, result := Translate(rec)

	body, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// Translate maps a recovered panic to the (status, body) pair the Java handler
// would produce. Split out from Handle so tests can assert on it directly.
func Translate(rec any) (int, *lang.Result) {
	switch e := rec.(type) {

	case *ShiroException:
		log.Printf("运行时异常：----------------%v", e)
		return http.StatusUnauthorized, lang.FailWith(401, e.GetMessage(), nil)

	case *MethodArgumentNotValidException:
		log.Printf("实体校验异常：----------------%v", e)
		// bindingResult.getAllErrors().stream().findFirst().get()
		return http.StatusBadRequest, lang.Fail(e.Errors[0])

	case *IllegalArgumentException:
		log.Printf("Assert异常：----------------%v", e)
		return http.StatusBadRequest, lang.Fail(e.GetMessage())

	case *NullPointerException:
		// A NullPointerException is a RuntimeException, so it lands in the
		// RuntimeException handler and calls Result.fail(e.getMessage()).
		// Under Java 8 that message is null -> "msg":null.
		log.Printf("运行时异常：----------------%v", e)
		return http.StatusBadRequest, lang.FailWithNullable(400, e.Message, nil)

	case *RuntimeException:
		log.Printf("运行时异常：----------------%v", e)
		return http.StatusBadRequest, lang.Fail(e.GetMessage())

	default:
		// Anything else is still a RuntimeException as far as Spring is
		// concerned.
		if err, ok := rec.(error); ok {
			log.Printf("运行时异常：----------------%v", err)
			return http.StatusBadRequest, lang.Fail(err.Error())
		}
		log.Printf("运行时异常：----------------%v", rec)
		return http.StatusBadRequest, lang.FailNullMsg()
	}
}
