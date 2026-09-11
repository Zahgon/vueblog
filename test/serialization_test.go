package test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
)

// TestBlogCreatedIsDateOnly pins @JsonFormat(pattern="yyyy-MM-dd") on
// Blog.created: the time component is dropped on the wire.
func TestBlogCreatedIsDateOnly(t *testing.T) {
	ts := time.Date(2020, 5, 21, 22, 8, 42, 0, time.Local)
	b := &entity.Blog{Created: lang.NewLocalDate(ts)}

	got, err := json.Marshal(b.Created)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(got) != `"2020-05-21"` {
		t.Errorf("Blog.created = %s, want \"2020-05-21\"", got)
	}
}

// TestUserCreatedIsIsoDateTime pins the unannotated LocalDateTime form. Spring
// Boot disables WRITE_DATES_AS_TIMESTAMPS, so it is ISO-8601 with no offset -
// not an epoch number and not a [y,m,d,...] array.
func TestUserCreatedIsIsoDateTime(t *testing.T) {
	ts := time.Date(2020, 4, 20, 10, 44, 1, 0, time.Local)
	u := &entity.User{Created: lang.NewLocalDateTime(ts)}

	got, err := json.Marshal(u.Created)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(got) != `"2020-04-20T10:44:01"` {
		t.Errorf("User.created = %s, want \"2020-04-20T10:44:01\"", got)
	}
}

// TestNullDatesSerialiseAsNull covers Jackson's ALWAYS inclusion for the
// nullable m_user.last_login column.
func TestNullDatesSerialiseAsNull(t *testing.T) {
	u := &entity.User{Id: ptr64(1), Username: "markerhub"}

	got, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"avatar", "email", "password", "status", "created", "lastLogin"} {
		v, present := decoded[key]
		if !present {
			t.Errorf("key %q is absent; Jackson ALWAYS inclusion must emit it", key)
		}
		if v != nil {
			t.Errorf("key %q = %v, want null", key, v)
		}
	}
}

// TestBlogFieldNamesAreCamelCase guards the JSON contract the Vue frontend
// consumes: MP maps user_id -> userId.
func TestBlogFieldNamesAreCamelCase(t *testing.T) {
	b := &entity.Blog{Id: ptr64(1), UserId: ptr64(2), Title: "t"}

	got, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := decoded["userId"]; !ok {
		t.Error("expected camelCase key userId")
	}
	if _, ok := decoded["user_id"]; ok {
		t.Error("snake_case user_id must not appear in the JSON")
	}
}

// TestAccessorsAreChainable mirrors @Accessors(chain = true).
func TestAccessorsAreChainable(t *testing.T) {
	b := (&entity.Blog{}).SetTitle("t").SetDescription("d").SetContent("c").SetStatus(0)

	if b.Title != "t" || b.Description != "d" || b.Content != "c" {
		t.Errorf("chained setters produced %+v", b)
	}
	if b.Status == nil || *b.Status != 0 {
		t.Error("SetStatus(0) did not take effect")
	}
}

func ptr64(v int64) *int64 { return &v }
