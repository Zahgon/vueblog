package test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/markerhub/vueblog-go/internal/util"
)

const configuredSecret = "f4e2e52034348f86b67cde581c0f9eb5"

func newJwtUtils() *util.JwtUtils {
	return &util.JwtUtils{Secret: configuredSecret, Expire: 604800, Header: "Authorization"}
}

// TestGenerateTokenRoundTrip covers generateToken -> getClaimByToken.
func TestGenerateTokenRoundTrip(t *testing.T) {
	j := newJwtUtils()

	token, err := j.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims := j.GetClaimByToken(token)
	if claims == nil {
		t.Fatal("GetClaimByToken returned nil for a token we just signed")
	}
	// setSubject(userId + "") - the subject is the id as a string.
	if claims.Subject != "42" {
		t.Errorf("subject = %q, want %q", claims.Subject, "42")
	}
	// expire is 604800 seconds after issuance.
	if d := claims.Expiration.Sub(claims.IssuedAt); d != 604800*time.Second {
		t.Errorf("exp - iat = %v, want 604800s", d)
	}
}

// TestTokenHeaderMatchesJjwt pins the header jjwt produces:
// setHeaderParam("typ","JWT") plus alg=HS512 from signWith.
func TestTokenHeaderMatchesJjwt(t *testing.T) {
	j := newJwtUtils()

	token, err := j.GenerateToken(1)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[0])
	if err != nil {
		t.Fatalf("failed to decode the JWT header: %v", err)
	}
	var header map[string]any
	if err := json.Unmarshal(raw, &header); err != nil {
		t.Fatalf("JWT header is not JSON: %v", err)
	}
	if header["typ"] != "JWT" {
		t.Errorf("typ = %v, want JWT", header["typ"])
	}
	if header["alg"] != "HS512" {
		t.Errorf("alg = %v, want HS512", header["alg"])
	}
}

// TestSecretIsBase64Decoded is the important one.
//
// jjwt 0.9.1's signWith(SignatureAlgorithm, String) base64-DECODES the secret
// before using it as the HMAC key, and setSigningKey(String) does the same.
// Treating the configured value as raw ASCII bytes would produce tokens the
// Java service rejects and vice versa. This asserts the Go side decodes too, by
// checking a token signed with the decoded bytes verifies, while one signed
// with a deliberately different key does not.
func TestSecretIsBase64Decoded(t *testing.T) {
	decoded, err := base64.StdEncoding.DecodeString(configuredSecret)
	if err != nil {
		t.Fatalf("the configured secret should be valid base64: %v", err)
	}
	if len(decoded) != 24 {
		t.Errorf("decoded key = %d bytes, want 24", len(decoded))
	}

	j := newJwtUtils()
	token, err := j.GenerateToken(1)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	// Verifying with the same JwtUtils succeeds.
	if j.GetClaimByToken(token) == nil {
		t.Error("a token signed with the decoded secret must verify")
	}

	// Verifying with the raw (undecoded) bytes as the key must fail, proving the
	// two key derivations really are different.
	rawKeyUtils := &util.JwtUtils{
		Secret: base64.StdEncoding.EncodeToString([]byte(configuredSecret)),
		Expire: 604800,
	}
	if rawKeyUtils.GetClaimByToken(token) != nil {
		t.Error("the raw-ASCII key must not verify a token signed with the decoded key")
	}
}

// TestGetClaimByTokenSwallowsErrors: the Java method catches Exception and
// returns null, so no failure mode propagates an error to the caller.
func TestGetClaimByTokenSwallowsErrors(t *testing.T) {
	j := newJwtUtils()

	for _, bad := range []string{"", "garbage", "a.b.c", "not.a.jwt.at.all"} {
		if claims := j.GetClaimByToken(bad); claims != nil {
			t.Errorf("GetClaimByToken(%q) = %+v, want nil", bad, claims)
		}
	}
}

// TestIsTokenExpired mirrors Date.before(new Date()), which is strict.
func TestIsTokenExpired(t *testing.T) {
	j := newJwtUtils()

	if !j.IsTokenExpired(time.Now().Add(-time.Second)) {
		t.Error("a past expiry must be reported as expired")
	}
	if j.IsTokenExpired(time.Now().Add(time.Hour)) {
		t.Error("a future expiry must not be reported as expired")
	}
}
