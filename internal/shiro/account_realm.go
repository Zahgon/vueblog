package shiro

import (
	"strconv"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/util"
)

// AccountRealm mirrors com.markerhub.shiro.AccountRealm extends AuthorizingRealm.
type AccountRealm struct {
	JwtUtils    *util.JwtUtils
	UserService service.UserService
}

func NewAccountRealm(j *util.JwtUtils, us service.UserService) *AccountRealm {
	return &AccountRealm{JwtUtils: j, UserService: us}
}

// Supports mirrors AccountRealm.supports(AuthenticationToken).
func (r *AccountRealm) Supports(token any) bool {
	_, ok := token.(*JwtToken)
	return ok
}

// DoGetAuthorizationInfo mirrors the override that returns null - vueblog does
// no role/permission checking.
func (r *AccountRealm) DoGetAuthorizationInfo() any { return nil }

// DoGetAuthenticationInfo mirrors AccountRealm.doGetAuthenticationInfo.
//
// Faithful detail: the Java code calls
//
//	jwtUtils.getClaimByToken(...).getSubject()
//
// with no null check. getClaimByToken returns null for a malformed or
// unverifiable token, so that line throws NullPointerException rather than a
// Shiro exception. The filter only reaches here after its own claim check, but
// the unguarded path is preserved so behaviour matches if it is ever hit.
func (r *AccountRealm) DoGetAuthenticationInfo(token *JwtToken) *AccountProfile {
	claims := r.JwtUtils.GetClaimByToken(token.GetPrincipal().(string))
	if claims == nil {
		exception.ThrowNPE()
	}

	userId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		// Long.valueOf(String) throws NumberFormatException, an
		// IllegalArgumentException subclass -> handled as HTTP 400.
		panic(&exception.IllegalArgumentException{
			Message: `For input string: "` + claims.Subject + `"`,
		})
	}

	user, err := r.UserService.GetById(userId)
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	if user == nil {
		panic(exception.NewUnknownAccount("账户不存在"))
	}

	// user.getStatus() == -1 auto-unboxes an Integer in Java, so a NULL status
	// column throws NPE. m_user.status is NOT NULL, but the unboxing path is
	// reproduced for exactness.
	if user.Status == nil {
		exception.ThrowNPE()
	}
	if *user.Status == -1 {
		panic(exception.NewLockedAccount("账户已被锁定"))
	}

	// BeanUtil.copyProperties(user, profile) copies by matching property name:
	// id, username, avatar and email exist on AccountProfile; password, status,
	// created and lastLogin have no counterpart and are dropped.
	return &AccountProfile{
		Id:       user.Id,
		Username: user.Username,
		Avatar:   user.Avatar,
		Email:    user.Email,
	}
}
