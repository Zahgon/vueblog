// Package validation reproduces the javax.validation constraints declared on
// LoginDto, User and Blog, including the exact message strings.
package validation

import (
	"regexp"
	"strings"

	"github.com/markerhub/vueblog-go/internal/common/dto"
	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/entity"
)

// notBlank mirrors @NotBlank: non-null and at least one non-whitespace char.
// Java's String.trim() strips code points <= U+0020, which differs from
// strings.TrimSpace (Unicode-aware); TrimFunc with the ASCII rule matches Java.
func notBlank(s string) bool {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' }) != ""
}

// emailPattern mirrors Hibernate Validator's @Email, which accepts any
// local-part@domain with no whitespace and a non-empty domain label.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)

func isEmail(s string) bool {
	if s == "" {
		return true // @Email alone permits null/empty; @NotBlank covers emptiness
	}
	return emailPattern.MatchString(s)
}

// ValidateLoginDto mirrors @Validated on AccountController.login.
// Field order follows the Java declaration order, which is the order
// Hibernate Validator reports and therefore what findFirst() picks.
func ValidateLoginDto(d *dto.LoginDto) {
	var errs []string
	if !notBlank(d.Username) {
		errs = append(errs, "昵称不能为空")
	}
	if !notBlank(d.Password) {
		errs = append(errs, "密码不能为空")
	}
	if len(errs) > 0 {
		panic(&exception.MethodArgumentNotValidException{Errors: errs})
	}
}

// ValidateUser mirrors @Validated on UserController.save.
func ValidateUser(u *entity.User) {
	var errs []string
	if !notBlank(u.Username) {
		errs = append(errs, "昵称不能为空")
	}
	email := ""
	if u.Email != nil {
		email = *u.Email
	}
	if !notBlank(email) {
		errs = append(errs, "邮箱不能为空")
	}
	if !isEmail(email) {
		errs = append(errs, "邮箱格式不正确")
	}
	if len(errs) > 0 {
		panic(&exception.MethodArgumentNotValidException{Errors: errs})
	}
}

// ValidateBlog mirrors @Validated on BlogController.edit.
func ValidateBlog(b *entity.Blog) {
	var errs []string
	if !notBlank(b.Title) {
		errs = append(errs, "标题不能为空")
	}
	if !notBlank(b.Description) {
		errs = append(errs, "摘要不能为空")
	}
	if !notBlank(b.Content) {
		errs = append(errs, "内容不能为空")
	}
	if len(errs) > 0 {
		panic(&exception.MethodArgumentNotValidException{Errors: errs})
	}
}
