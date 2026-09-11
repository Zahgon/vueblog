package shiro

import "context"

// GetProfile mirrors com.markerhub.util.ShiroUtil.getProfile():
//
//	return (AccountProfile) SecurityUtils.getSubject().getPrincipal();
//
// Structural note: the Java class lives in com.markerhub.util, but util already
// holds JwtUtils which the shiro package imports. Putting ShiroUtil in util
// would create an import cycle (util -> shiro -> util), which Go forbids, so it
// lives beside the AccountProfile it returns. This is the one place the Go
// package layout departs from the Java package layout.
//
// Behaviour is unchanged: an anonymous subject yields nil, and callers that
// dereference it reproduce the Java NullPointerException.
func GetProfile(ctx context.Context) *AccountProfile {
	return GetSubject(ctx).GetPrincipal()
}
