// Package controller mirrors com.markerhub.controller.
package controller

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/markerhub/vueblog-go/internal/common/dto"
	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/shiro"
	"github.com/markerhub/vueblog-go/internal/util"
	"github.com/markerhub/vueblog-go/internal/validation"
)

// AccountController mirrors com.markerhub.controller.AccountController.
type AccountController struct {
	UserService service.UserService
	JwtUtils    *util.JwtUtils
}

func NewAccountController(us service.UserService, j *util.JwtUtils) *AccountController {
	return &AccountController{UserService: us, JwtUtils: j}
}

// md5Hex mirrors cn.hutool.crypto.SecureUtil.md5(String): lowercase hex of the
// UTF-8 bytes.
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// loginResponse mirrors the hutool MapUtil.builder() map returned on success.
//
// Java builds a HashMap, whose JSON key order is unspecified. A struct gives a
// deterministic order matching the Java put() order; JSON object key order is
// not semantically significant, so this is behaviour-preserving.
type loginResponse struct {
	Id       *int64  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Email    *string `json:"email"`
}

// Login mirrors @PostMapping("/login").
func (c *AccountController) Login(w http.ResponseWriter, r *http.Request) *lang.Result {
	var loginDto dto.LoginDto
	if err := json.NewDecoder(r.Body).Decode(&loginDto); err != nil {
		// A malformed body raises HttpMessageNotReadableException in Spring,
		// which this app does not handle explicitly -> RuntimeException -> 400.
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	validation.ValidateLoginDto(&loginDto)

	user, err := c.UserService.GetOne(mybatisplus.NewQueryWrapper().Eq("username", loginDto.Username))
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	exception.NotNull(user, "用户不存在")

	// user.getPassword().equals(...) - a NULL password column would NPE here in
	// Java, so the nil case is reproduced rather than treated as a mismatch.
	if user.Password == nil {
		exception.ThrowNPE()
	}
	if *user.Password != md5Hex(loginDto.Password) {
		// Returned, not thrown: HTTP stays 200 with code 400 in the body.
		return lang.Fail("密码不正确")
	}

	jwt, err := c.JwtUtils.GenerateToken(*user.Id)
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}

	w.Header().Set("Authorization", jwt)
	w.Header().Set("Access-control-Expose-Headers", "Authorization")

	return lang.Succ(&loginResponse{
		Id:       user.Id,
		Username: user.Username,
		Avatar:   user.Avatar,
		Email:    user.Email,
	})
}

// Logout mirrors @RequiresAuthentication @GetMapping("/logout").
func (c *AccountController) Logout(w http.ResponseWriter, r *http.Request) *lang.Result {
	RequiresAuthentication(r)
	shiro.GetSubject(r.Context()).Logout()
	return lang.Succ(nil)
}
