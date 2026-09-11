// Package dto mirrors com.markerhub.common.dto.
package dto

// LoginDto mirrors com.markerhub.common.dto.LoginDto.
// The validation constraints are declared in validation.LoginDtoRules so the
// message text and evaluation order stay identical to javax.validation's.
type LoginDto struct {
	Username string `json:"username"` // @NotBlank(message = "昵称不能为空")
	Password string `json:"password"` // @NotBlank(message = "密码不能为空")
}
