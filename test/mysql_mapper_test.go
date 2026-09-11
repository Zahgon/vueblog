package test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mapper"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// The production mappers are what MyBatis-Plus generated for BaseMapper<Blog>
// and BaseMapper<User>. In Java they were never unit-tested because the
// framework supplied them; here they are the port's own code, so the SQL they
// emit and the way they map rows back are pinned against a mock driver.

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("opening sqlmock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
		_ = db.Close()
	})
	return db, mock
}

func blogRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "title", "description", "content", "created", "status"})
}

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "avatar", "email", "password", "status", "created", "last_login"})
}

func TestBlogMapperSelectById(t *testing.T) {
	db, mock := newMock(t)
	created := time.Date(2020, 4, 12, 15, 4, 5, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, description, content, created, status FROM m_blog WHERE id = ?")).
		WithArgs(int64(7)).
		WillReturnRows(blogRows().AddRow(7, 1, "t", "d", "c", created, 0))

	blog, err := mapper.NewMySQLBlogMapper(db).SelectById(7)
	if err != nil {
		t.Fatalf("SelectById: %v", err)
	}
	if blog == nil || blog.Id == nil || *blog.Id != 7 {
		t.Fatalf("SelectById = %+v, want id 7", blog)
	}
	if blog.Title != "t" || blog.Description != "d" || blog.Content != "c" {
		t.Errorf("row mapped as %+v", blog)
	}
	if blog.Created == nil || !blog.Created.T.Equal(created) {
		t.Errorf("Created = %v, want %v", blog.Created, created)
	}
}

func TestBlogMapperSelectByIdMissingReturnsNilNotError(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_blog WHERE id = ?")).
		WithArgs(int64(404)).
		WillReturnError(sql.ErrNoRows)

	// MyBatis returns null for a missing row rather than raising.
	blog, err := mapper.NewMySQLBlogMapper(db).SelectById(404)
	if err != nil {
		t.Fatalf("SelectById = %v, want nil error", err)
	}
	if blog != nil {
		t.Errorf("SelectById = %+v, want nil", blog)
	}
}

func TestBlogMapperSelectByIdPropagatesRealErrors(t *testing.T) {
	db, mock := newMock(t)
	boom := errors.New("connection lost")
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_blog WHERE id = ?")).
		WithArgs(int64(1)).
		WillReturnError(boom)

	if _, err := mapper.NewMySQLBlogMapper(db).SelectById(1); err == nil {
		t.Error("SelectById swallowed a driver error")
	}
}

func TestBlogMapperSelectPageCountsThenPages(t *testing.T) {
	db, mock := newMock(t)
	page := mybatisplus.NewPage[*entity.Blog](2, 5)
	qw := mybatisplus.NewQueryWrapper().Eq("status", 0)

	// PaginationInterceptor issues the COUNT first when isSearchCount is true.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_blog WHERE status = ?")).
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_blog WHERE status = ?")).
		WithArgs(0, int64(5), int64(5)).
		WillReturnRows(blogRows().
			AddRow(6, 1, "six", "d6", "c6", nil, 0).
			AddRow(7, 1, "seven", "d7", "c7", nil, 0))

	if err := mapper.NewMySQLBlogMapper(db).SelectPage(page, qw); err != nil {
		t.Fatalf("SelectPage: %v", err)
	}
	if page.Total != 11 {
		t.Errorf("Total = %d, want 11", page.Total)
	}
	if len(page.Records) != 2 {
		t.Fatalf("Records = %d, want 2", len(page.Records))
	}
	if page.Records[0].Title != "six" || page.Records[1].Title != "seven" {
		t.Errorf("records out of order: %+v", page.Records)
	}
}

func TestBlogMapperSelectPageSkipsCountWhenNotSearching(t *testing.T) {
	db, mock := newMock(t)
	page := mybatisplus.NewPage[*entity.Blog](1, 10)
	page.SearchCount = false
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_blog")).
		WillReturnRows(blogRows())

	if err := mapper.NewMySQLBlogMapper(db).SelectPage(page, mybatisplus.NewQueryWrapper()); err != nil {
		t.Fatalf("SelectPage: %v", err)
	}
	// An empty page still carries a non-nil, empty record list.
	if page.Records == nil || len(page.Records) != 0 {
		t.Errorf("Records = %+v, want an empty non-nil slice", page.Records)
	}
}

func TestBlogMapperInsertWritesTheGeneratedKeyBack(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO m_blog (user_id, title, description, content, created, status) VALUES (?, ?, ?, ?, ?, ?)")).
		WillReturnResult(sqlmock.NewResult(42, 1))

	blog := (&entity.Blog{}).SetTitle("t").SetDescription("d").SetContent("c")
	if err := mapper.NewMySQLBlogMapper(db).Insert(blog); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	// IdType.AUTO writes the generated key back onto the entity.
	if blog.Id == nil || *blog.Id != 42 {
		t.Errorf("Id = %v, want 42", blog.Id)
	}
}

func TestBlogMapperUpdateById(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE m_blog SET user_id = ?, title = ?, description = ?, content = ?, created = ?, status = ? WHERE id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	blog := (&entity.Blog{}).SetId(3).SetTitle("t2")
	if err := mapper.NewMySQLBlogMapper(db).UpdateById(blog); err != nil {
		t.Fatalf("UpdateById: %v", err)
	}
}

func TestUserMapperSelectById(t *testing.T) {
	db, mock := newMock(t)
	created := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, avatar, email, password, status, created, last_login FROM m_user WHERE id = ?")).
		WithArgs(int64(1)).
		WillReturnRows(userRows().AddRow(1, "mark", "a.png", "m@e.com", "pw", 0, created, created))

	user, err := mapper.NewMySQLUserMapper(db).SelectById(1)
	if err != nil {
		t.Fatalf("SelectById: %v", err)
	}
	if user == nil || user.Username != "mark" {
		t.Fatalf("SelectById = %+v", user)
	}
	if user.Email == nil || *user.Email != "m@e.com" {
		t.Errorf("Email = %v", user.Email)
	}
	if user.LastLogin == nil || !user.LastLogin.T.Equal(created) {
		t.Errorf("LastLogin = %v", user.LastLogin)
	}
}

func TestUserMapperSelectByIdNullableColumnsStayAbsent(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_user WHERE id = ?")).
		WithArgs(int64(2)).
		WillReturnRows(userRows().AddRow(2, "n", nil, nil, nil, nil, nil, nil))

	user, err := mapper.NewMySQLUserMapper(db).SelectById(2)
	if err != nil {
		t.Fatalf("SelectById: %v", err)
	}
	// A NULL column leaves the pointer nil rather than zeroing it.
	if user.Avatar != nil || user.Email != nil || user.Password != nil ||
		user.Status != nil || user.Created != nil || user.LastLogin != nil {
		t.Errorf("NULL columns were materialised: %+v", user)
	}
}

func TestUserMapperSelectOne(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_user WHERE username = ?")).
		WithArgs("mark").
		WillReturnRows(userRows().AddRow(1, "mark", nil, nil, "pw", 0, nil, nil))

	user, err := mapper.NewMySQLUserMapper(db).SelectOne(
		mybatisplus.NewQueryWrapper().Eq("username", "mark"),
	)
	if err != nil {
		t.Fatalf("SelectOne: %v", err)
	}
	if user == nil || user.Username != "mark" {
		t.Fatalf("SelectOne = %+v", user)
	}
}

func TestUserMapperSelectOneMissingReturnsNil(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM m_user WHERE username = ?")).
		WithArgs("absent").
		WillReturnError(sql.ErrNoRows)

	user, err := mapper.NewMySQLUserMapper(db).SelectOne(
		mybatisplus.NewQueryWrapper().Eq("username", "absent"),
	)
	if err != nil {
		t.Fatalf("SelectOne = %v, want nil error", err)
	}
	if user != nil {
		t.Errorf("SelectOne = %+v, want nil", user)
	}
}
