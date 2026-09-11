// Package test mirrors src/test/java/com/markerhub.
package test

import (
	"testing"

	"github.com/markerhub/vueblog-go/internal/testsupport"
)

// TestContextLoads is the direct migration of
// src/test/java/com/markerhub/VueblogApplicationTests.java:
//
//	@SpringBootTest
//	class VueblogApplicationTests {
//	    @Test
//	    void contextLoads() {
//	    }
//	}
//
// The Java test body is empty: @SpringBootTest builds the entire application
// context and the test passes if - and only if - that wiring succeeds. The Go
// equivalent is to construct the same object graph and assert every bean was
// created, which is what testsupport.New does.
func TestContextLoads(t *testing.T) {
	app := testsupport.New()

	if app.Handler == nil {
		t.Fatal("application context failed to load: no dispatcher handler")
	}
	if app.JwtUtils == nil {
		t.Fatal("application context failed to load: JwtUtils bean missing")
	}
	if app.UserService == nil {
		t.Fatal("application context failed to load: UserService bean missing")
	}
	if app.BlogService == nil {
		t.Fatal("application context failed to load: BlogService bean missing")
	}
}
