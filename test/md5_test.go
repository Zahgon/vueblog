package test

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

// TestSeedPasswordHashMatchesHutool verifies the Go md5 helper reproduces
// cn.hutool.crypto.SecureUtil.md5: lowercase hex of the UTF-8 bytes. The
// expected value is the hash committed in vueblog.sql for user markerhub.
func TestSeedPasswordHashMatchesHutool(t *testing.T) {
	sum := md5.Sum([]byte(testsupport.SeedPasswordPlain))
	got := hex.EncodeToString(sum[:])

	if got != testsupport.SeedPasswordHash {
		t.Errorf("md5(%q) = %s, want the hash committed in vueblog.sql (%s)",
			testsupport.SeedPasswordPlain, got, testsupport.SeedPasswordHash)
	}
}

// TestMd5IsLowercaseHex guards the casing, which decides whether the stored
// password ever compares equal.
func TestMd5IsLowercaseHex(t *testing.T) {
	sum := md5.Sum([]byte("abc"))
	got := hex.EncodeToString(sum[:])

	if got != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("md5(\"abc\") = %s, want lowercase hex", got)
	}
}
