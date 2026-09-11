package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/markerhub/vueblog-go/internal/config"
	"github.com/markerhub/vueblog-go/internal/entity"
)

// TestDefaultsMatchApplicationYml pins the committed application.yml values, so
// a service started without a config file behaves like the Java default profile.
func TestDefaultsMatchApplicationYml(t *testing.T) {
	p := config.Defaults()

	if p.Server.Port != 8081 {
		t.Errorf("server.port = %d, want 8081", p.Server.Port)
	}
	if p.Markerhub.Jwt.Secret != "f4e2e52034348f86b67cde581c0f9eb5" {
		t.Errorf("markerhub.jwt.secret = %q", p.Markerhub.Jwt.Secret)
	}
	if p.Markerhub.Jwt.Expire != 604800 {
		t.Errorf("markerhub.jwt.expire = %d, want 604800", p.Markerhub.Jwt.Expire)
	}
	if p.Markerhub.Jwt.Header != "Authorization" {
		t.Errorf("markerhub.jwt.header = %q, want Authorization", p.Markerhub.Jwt.Header)
	}
	if !p.ShiroRedis.Enabled || p.ShiroRedis.RedisManager.Host != "127.0.0.1:6379" {
		t.Errorf("shiro-redis = %+v, want enabled on 127.0.0.1:6379", p.ShiroRedis)
	}
}

// TestLoadOverridesDefaults covers YAML binding.
func TestLoadOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "application.yml")
	yaml := "server:\n  port: 9999\nmarkerhub:\n  jwt:\n    expire: 60\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	p, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Server.Port != 9999 {
		t.Errorf("overridden port = %d, want 9999", p.Server.Port)
	}
	if p.Markerhub.Jwt.Expire != 60 {
		t.Errorf("overridden expire = %d, want 60", p.Markerhub.Jwt.Expire)
	}
	// Unspecified keys keep their defaults.
	if p.Markerhub.Jwt.Secret != "f4e2e52034348f86b67cde581c0f9eb5" {
		t.Error("an unspecified key must retain its default")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := config.Load(filepath.Join(t.TempDir(), "nope.yml")); err == nil {
		t.Error("Load on a missing file must return an error")
	}
}

// TestFilterChainDefinition mirrors ShiroConfig: every path runs the jwt filter.
func TestFilterChainDefinition(t *testing.T) {
	if config.FilterChainDefinition["/**"] != "jwt" {
		t.Errorf("filter chain = %v, want /** -> jwt", config.FilterChainDefinition)
	}
}

// TestUserAccessorsAreChainable mirrors @Accessors(chain = true) on User.
func TestUserAccessorsAreChainable(t *testing.T) {
	u := (&entity.User{}).
		SetId(7).
		SetUsername("markerhub").
		SetAvatar("a.jpg").
		SetEmail("a@b.com").
		SetPassword("hash").
		SetStatus(0)

	if u.Id == nil || *u.Id != 7 {
		t.Error("SetId did not take effect")
	}
	if u.Username != "markerhub" {
		t.Errorf("username = %q", u.Username)
	}
	if u.Avatar == nil || *u.Avatar != "a.jpg" {
		t.Error("SetAvatar did not take effect")
	}
	if u.Email == nil || *u.Email != "a@b.com" {
		t.Error("SetEmail did not take effect")
	}
	if u.Password == nil || *u.Password != "hash" {
		t.Error("SetPassword did not take effect")
	}
	if u.GetStatus() == nil || *u.GetStatus() != 0 {
		t.Error("SetStatus/GetStatus did not round-trip")
	}
	if entity.UserTableName != "m_user" || entity.BlogTableName != "m_blog" {
		t.Error("@TableName values must match the DDL in vueblog.sql")
	}
}

// TestNotFound covers the unmapped-path branch.
func TestNotFound(t *testing.T) {
	app := newAppForNotFound(t)
	req := do(t, app, "GET", "/does/not/exist", "", "")
	if req.Status != 404 {
		t.Errorf("status = %d, want 404", req.Status)
	}
}
