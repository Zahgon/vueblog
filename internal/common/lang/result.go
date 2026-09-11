// Package lang mirrors com.markerhub.common.lang.
package lang

// Result mirrors com.markerhub.common.lang.Result (@Data, Serializable).
//
// Jackson serialises the Lombok-generated getters in field-declaration order,
// so the JSON key order is code, msg, data - matching the Java output.
// Jackson's default inclusion is ALWAYS (this app never sets
// serializationInclusion), so nil fields must still emit null.
//
// Msg is a *string rather than a string because Result.fail(String) is reachable
// with a null argument: GlobalExceptionHandler.handler(RuntimeException) calls
// Result.fail(e.getMessage()), and a Java 8 NullPointerException has a null
// message. That path must serialise "msg":null, not "msg":"".
type Result struct {
	Code int     `json:"code"` // 200是正常，非200表示异常
	Msg  *string `json:"msg"`
	Data any     `json:"data"`
}

// GetMsg returns the message, or "" when it is null. Convenience for callers
// that do not care about the null case.
func (r *Result) GetMsg() string {
	if r.Msg == nil {
		return ""
	}
	return *r.Msg
}

func str(s string) *string { return &s }

// Succ mirrors Result.succ(Object data).
func Succ(data any) *Result {
	return SuccWith(200, "操作成功", data)
}

// SuccWith mirrors Result.succ(int code, String msg, Object data).
func SuccWith(code int, msg string, data any) *Result {
	return &Result{Code: code, Msg: str(msg), Data: data}
}

// Fail mirrors Result.fail(String msg).
func Fail(msg string) *Result {
	return FailWith(400, msg, nil)
}

// FailNullMsg mirrors Result.fail(null) - reachable when a RuntimeException
// carries no message.
func FailNullMsg() *Result {
	return &Result{Code: 400, Msg: nil, Data: nil}
}

// FailData mirrors Result.fail(String msg, Object data).
func FailData(msg string, data any) *Result {
	return FailWith(400, msg, data)
}

// FailWith mirrors Result.fail(int code, String msg, Object data).
func FailWith(code int, msg string, data any) *Result {
	return &Result{Code: code, Msg: str(msg), Data: data}
}

// FailWithNullable mirrors Result.fail(int, String, Object) where the message
// may itself be null.
func FailWithNullable(code int, msg *string, data any) *Result {
	return &Result{Code: code, Msg: msg, Data: data}
}
