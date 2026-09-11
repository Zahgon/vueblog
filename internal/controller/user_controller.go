package controller

import (
	"encoding/json"
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/validation"
)

// UserController mirrors com.markerhub.controller.UserController
// (@RequestMapping("/user")).
type UserController struct{ UserService service.UserService }

func NewUserController(us service.UserService) *UserController {
	return &UserController{UserService: us}
}

// Index mirrors @RequiresAuthentication @GetMapping("/index").
// The Java code hardcodes getById(1L) - it does not read the logged-in user.
func (c *UserController) Index(w http.ResponseWriter, r *http.Request) *lang.Result {
	RequiresAuthentication(r)

	user, err := c.UserService.GetById(1)
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	// No null check in the original: a missing row simply yields "data":null.
	return lang.Succ(user)
}

// Save mirrors @PostMapping("/save"). It validates and echoes the user back
// without persisting anything - matching the original, which has no save call.
func (c *UserController) Save(w http.ResponseWriter, r *http.Request) *lang.Result {
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	validation.ValidateUser(&user)
	return lang.Succ(&user)
}
