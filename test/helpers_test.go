package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

// response is the decoded Result envelope.
type response struct {
	Status int
	Code   int             `json:"code"`
	Msg    *string         `json:"msg"`
	Data   json.RawMessage `json:"data"`
	Header http.Header
	Raw    string
}

func (r response) msg() string {
	if r.Msg == nil {
		return "<null>"
	}
	return *r.Msg
}

// do issues a request against the wired application.
func do(t *testing.T, app *testsupport.App, method, path, body, token string) response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	res := response{Status: rec.Code, Header: rec.Header(), Raw: rec.Body.String()}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("%s %s: response is not a Result envelope: %v\nbody: %s",
				method, path, err, rec.Body.String())
		}
	}
	return res
}

func assertResult(t *testing.T, got response, wantStatus, wantCode int, wantMsg string) {
	t.Helper()
	if got.Status != wantStatus {
		t.Errorf("HTTP status = %d, want %d (body: %s)", got.Status, wantStatus, got.Raw)
	}
	if got.Code != wantCode {
		t.Errorf("Result.code = %d, want %d (body: %s)", got.Code, wantCode, got.Raw)
	}
	if got.msg() != wantMsg {
		t.Errorf("Result.msg = %q, want %q", got.msg(), wantMsg)
	}
}
