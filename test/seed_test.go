package test

import (
	"testing"

	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/testsupport"
	"time"
)

func seedUser(id int64, username string) *entity.User {
	status := int32(0)
	pass := "96e79218965eb72c92a549dd5a330112"
	return &entity.User{
		Id:       &id,
		Username: username,
		Password: &pass,
		Status:   &status,
		Created:  lang.NewLocalDateTime(time.Now()),
	}
}

// appAlias keeps the helper signature readable.
type appAlias = testsupport.App

func newAppForNotFound(t *testing.T) *appAlias {
	t.Helper()
	return testsupport.New()
}
