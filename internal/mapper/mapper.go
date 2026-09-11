// Package mapper mirrors com.markerhub.mapper - the MyBatis-Plus BaseMapper
// interfaces. Two implementations are provided: MySQL (production, matching the
// original JDBC behaviour) and in-memory (used by the tests, which is what lets
// the migrated suite run without a live MySQL server).
package mapper

import (
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// BlogMapper mirrors com.markerhub.mapper.BlogMapper extends BaseMapper<Blog>.
type BlogMapper interface {
	SelectById(id int64) (*entity.Blog, error)
	SelectPage(page *mybatisplus.Page[*entity.Blog], qw *mybatisplus.QueryWrapper) error
	Insert(b *entity.Blog) error
	UpdateById(b *entity.Blog) error
}

// UserMapper mirrors com.markerhub.mapper.UserMapper extends BaseMapper<User>.
type UserMapper interface {
	SelectById(id int64) (*entity.User, error)
	SelectOne(qw *mybatisplus.QueryWrapper) (*entity.User, error)
}
