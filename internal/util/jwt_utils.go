// Package util mirrors com.markerhub.util.
package util

import (
	"encoding/base64"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtUtils mirrors com.markerhub.util.JwtUtils, bound to the markerhub.jwt
// prefix in application.yml.
type JwtUtils struct {
	Secret string
	Expire int64 // seconds
	Header string
}

// signingKey reproduces a subtle jjwt 0.9.1 behaviour that is easy to miss.
//
// JwtBuilder.signWith(SignatureAlgorithm, String) does NOT use the string's raw
// bytes as the HMAC key - it calls TextCodec.BASE64.decode(secret) first, and
// Jwts.parser().setSigningKey(String) does the same. So the configured secret
// "f4e2e52034348f86b67cde581c0f9eb5" is base64-decoded to 24 key bytes rather
// than used as 32 ASCII bytes. Signing with the raw string would produce tokens
// the Java service rejects, so we decode identically here.
//
// jjwt uses a lenient decoder, so we fall back to the raw bytes if the secret
// is not valid base64 - matching TextCodec's behaviour of never failing hard on
// a well-formed-looking key.
func (j *JwtUtils) signingKey() []byte {
	if b, err := base64.StdEncoding.DecodeString(j.Secret); err == nil {
		return b
	}
	if b, err := base64.RawStdEncoding.DecodeString(j.Secret); err == nil {
		return b
	}
	return []byte(j.Secret)
}

// GenerateToken mirrors JwtUtils.generateToken(long userId).
// jjwt writes iat/exp as NumericDate (seconds), sets typ=JWT in the header and
// alg=HS512 automatically.
func (j *JwtUtils) GenerateToken(userId int64) (string, error) {
	nowDate := time.Now()
	expireDate := nowDate.Add(time.Duration(j.Expire) * time.Second)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": strconv.FormatInt(userId, 10),
		"iat": nowDate.Unix(),
		"exp": expireDate.Unix(),
	})
	token.Header["typ"] = "JWT"
	return token.SignedString(j.signingKey())
}

// Claims mirrors the subset of io.jsonwebtoken.Claims this app reads.
type Claims struct {
	Subject    string
	IssuedAt   time.Time
	Expiration time.Time
}

// GetClaimByToken mirrors JwtUtils.getClaimByToken(String).
// The Java method swallows every exception and returns null, so this returns
// nil rather than an error - callers branch on nil exactly as the Java does.
//
// jjwt 0.9.1's parseClaimsJws rejects an expired token by throwing
// ExpiredJwtException, which this catch block turns into null. We therefore
// disable the library's own expiry check and let the caller apply
// IsTokenExpired, keeping the null-vs-expired distinction identical.
func (j *JwtUtils) GetClaimByToken(token string) *Claims {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return j.signingKey(), nil
	}, jwt.WithValidMethods([]string{"HS512"}), jwt.WithoutClaimsValidation())
	if err != nil || !parsed.Valid {
		log.Printf("validate is token error %v", err)
		return nil
	}
	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	c := &Claims{}
	if sub, ok := mc["sub"].(string); ok {
		c.Subject = sub
	}
	if iat, ok := mc["iat"].(float64); ok {
		c.IssuedAt = time.Unix(int64(iat), 0)
	}
	if exp, ok := mc["exp"].(float64); ok {
		c.Expiration = time.Unix(int64(exp), 0)
	}
	return c
}

// IsTokenExpired mirrors JwtUtils.isTokenExpired(Date): true when expired.
// Date.before is strict, so an expiry exactly equal to now is NOT expired.
func (j *JwtUtils) IsTokenExpired(expiration time.Time) bool {
	return expiration.Before(time.Now())
}
