package test

import (
	"encoding/json"
	"testing"

	"github.com/markerhub/vueblog-go/internal/common/lang"
)

// TestResultSuccJSON pins the exact envelope Jackson emits, including key order
// (field declaration order) and the null data field under ALWAYS inclusion.
func TestResultSuccJSON(t *testing.T) {
	got, err := json.Marshal(lang.Succ(nil))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"code":200,"msg":"操作成功","data":null}`
	if string(got) != want {
		t.Errorf("Result.succ(null) = %s\nwant %s", got, want)
	}
}

func TestResultFailJSON(t *testing.T) {
	got, err := json.Marshal(lang.Fail("用户不存在"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"code":400,"msg":"用户不存在","data":null}`
	if string(got) != want {
		t.Errorf("Result.fail(msg) = %s\nwant %s", got, want)
	}
}

// TestResultFailNullMsgJSON covers Result.fail(null), reachable when a
// RuntimeException carries no message - "msg" must be JSON null, not "".
func TestResultFailNullMsgJSON(t *testing.T) {
	got, err := json.Marshal(lang.FailNullMsg())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"code":400,"msg":null,"data":null}`
	if string(got) != want {
		t.Errorf("Result.fail(null) = %s\nwant %s", got, want)
	}
}

// TestResultCodes pins the static factory defaults.
func TestResultCodes(t *testing.T) {
	if r := lang.Succ("x"); r.Code != 200 || r.GetMsg() != "操作成功" {
		t.Errorf("succ = code %d msg %q, want 200 / 操作成功", r.Code, r.GetMsg())
	}
	if r := lang.Fail("x"); r.Code != 400 {
		t.Errorf("fail code = %d, want 400", r.Code)
	}
	if r := lang.FailWith(401, "x", nil); r.Code != 401 {
		t.Errorf("fail(401,..) code = %d, want 401", r.Code)
	}
	if r := lang.FailData("x", 7); r.Code != 400 || r.Data != 7 {
		t.Errorf("fail(msg,data) = %+v, want code 400 data 7", r)
	}
}
