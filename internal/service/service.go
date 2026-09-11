// Package service mirrors com.markerhub.service and .service.impl.
package service

import (
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mapper"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
)

// BlogService mirrors com.markerhub.service.BlogService extends IService<Blog>,
// implemented by BlogServiceImpl extends ServiceImpl<BlogMapper, Blog>.
// Only the IService methods vueblog actually calls are reproduced.
type BlogService interface {
	GetById(id int64) (*entity.Blog, error)
	Page(page *mybatisplus.Page[*entity.Blog], qw *mybatisplus.QueryWrapper) (*mybatisplus.Page[*entity.Blog], error)
	SaveOrUpdate(b *entity.Blog) (bool, error)
}

// UserService mirrors com.markerhub.service.UserService extends IService<User>.
type UserService interface {
	GetById(id int64) (*entity.User, error)
	GetOne(qw *mybatisplus.QueryWrapper) (*entity.User, error)
}

// BlogServiceImpl mirrors com.markerhub.service.impl.BlogServiceImpl.
type BlogServiceImpl struct{ Mapper mapper.BlogMapper }

func NewBlogService(m mapper.BlogMapper) *BlogServiceImpl { return &BlogServiceImpl{Mapper: m} }

func (s *BlogServiceImpl) GetById(id int64) (*entity.Blog, error) {
	return s.Mapper.SelectById(id)
}

func (s *BlogServiceImpl) Page(page *mybatisplus.Page[*entity.Blog], qw *mybatisplus.QueryWrapper) (*mybatisplus.Page[*entity.Blog], error) {
	if err := s.Mapper.SelectPage(page, qw); err != nil {
		return nil, err
	}
	return page, nil
}

// SaveOrUpdate mirrors IService.saveOrUpdate(T): insert when the id is null or
// no row with that id exists, otherwise update. That exact rule is what MP's
// ServiceImpl.saveOrUpdate implements.
func (s *BlogServiceImpl) SaveOrUpdate(b *entity.Blog) (bool, error) {
	if b == nil {
		return false, nil
	}
	if b.Id == nil {
		return true, s.Mapper.Insert(b)
	}
	existing, err := s.Mapper.SelectById(*b.Id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return true, s.Mapper.Insert(b)
	}
	return true, s.Mapper.UpdateById(b)
}

// UserServiceImpl mirrors com.markerhub.service.impl.UserServiceImpl.
type UserServiceImpl struct{ Mapper mapper.UserMapper }

func NewUserService(m mapper.UserMapper) *UserServiceImpl { return &UserServiceImpl{Mapper: m} }

func (s *UserServiceImpl) GetById(id int64) (*entity.User, error) {
	return s.Mapper.SelectById(id)
}

func (s *UserServiceImpl) GetOne(qw *mybatisplus.QueryWrapper) (*entity.User, error) {
	return s.Mapper.SelectOne(qw)
}
