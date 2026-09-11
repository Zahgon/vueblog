// Package testsupport builds a fully wired application backed by in-memory
// mappers seeded from resources/vueblog.sql.
//
// The Java suite could only run against a live MySQL + Redis pair, which is why
// it contained a single context-load test. Substituting the mapper layer keeps
// every other component - filter, realm, controllers, exception handler,
// serialisation - exactly as it runs in production, so the migrated tests
// exercise real behaviour deterministically.
package testsupport

import (
	"net/http"
	"time"

	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/config"
	"github.com/markerhub/vueblog-go/internal/controller"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/shiro"
	"github.com/markerhub/vueblog-go/internal/util"
)

// App is the assembled application context.
type App struct {
	Handler     http.Handler
	JwtUtils    *util.JwtUtils
	UserService service.UserService
	BlogService service.BlogService
	UserMapper  *MemoryUserMapper
	BlogMapper  *MemoryBlogMapper
}

// SeedPasswordPlain is the plaintext behind the md5 hash committed in
// vueblog.sql for user markerhub: 96e79218965eb72c92a549dd5a330112.
const SeedPasswordPlain = "111111"

// SeedPasswordHash is that committed hash.
const SeedPasswordHash = "96e79218965eb72c92a549dd5a330112"

func ptrInt64(v int64) *int64    { return &v }
func ptrInt32(v int32) *int32    { return &v }
func ptrString(v string) *string { return &v }

func mustTime(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

// New builds the app with the seed rows from vueblog.sql.
func New() *App {
	um := NewMemoryUserMapper()
	bm := NewMemoryBlogMapper()

	// INSERT INTO `m_user` VALUES ('1','markerhub','https://...jpg', null,
	//   '96e79218965eb72c92a549dd5a330112','0','2020-04-20 10:44:01', null);
	um.Seed(&entity.User{
		Id:       ptrInt64(1),
		Username: "markerhub",
		Avatar:   ptrString("https://image-1300566513.cos.ap-guangzhou.myqcloud.com/upload/images/5a9f48118166308daba8b6da7e466aab.jpg"),
		Email:    nil,
		Password: ptrString(SeedPasswordHash),
		Status:   ptrInt32(0),
		Created:  lang.NewLocalDateTime(mustTime("2020-04-20 10:44:01")),
	})

	// The six m_blog rows from vueblog.sql, all owned by user 1.
	seedBlog := func(id int64, title, desc, content, created string) {
		bm.Seed(&entity.Blog{
			Id:          ptrInt64(id),
			UserId:      ptrInt64(1),
			Title:       title,
			Description: desc,
			Content:     content,
			Created:     lang.NewLocalDate(mustTime(created)),
			Status:      ptrInt32(0),
		})
	}
	seedBlog(1, "生活就像海洋，只有意志坚强的人才能到达彼岸", "这里是摘要哈哈哈", "内容？？？", "2020-05-21 22:08:42")
	seedBlog(2, "最值得学习的博客项目eblog", "eblog是一个基于Springboot2.1.2开发的博客学习项目", "**推荐阅读：**", "2020-05-28 09:36:38")
	seedBlog(3, "关注公众号JavaCat，回复xshell或navicat获取破解对应工具", "视频中所用到的xshell和navicat直接获取哈！", "### 工具获取", "2020-10-20 05:05:31")
	seedBlog(7, "你真的会写单例模式吗?", "单例模式可能是代码最少的模式了", "饿汉法", "2020-05-22 00:42:44")
	seedBlog(9, "真正理解Mysql的四种隔离级别@", "事务是应用程序中一系列严密的操作", "### 什么是事务", "2020-05-22 22:04:46")
	seedBlog(10, "博客项目eblog讲解视频上线啦，长达17个小时！！", "1. 慕课网免费资源好久都没更新了", "ok，再回到我们的eblog项目", "2020-05-22 22:05:49")

	// application.yml: markerhub.jwt.*
	jwtUtils := &util.JwtUtils{
		Secret: "f4e2e52034348f86b67cde581c0f9eb5",
		Expire: 604800,
		Header: "Authorization",
	}

	userService := service.NewUserService(um)
	blogService := service.NewBlogService(bm)
	realm := shiro.NewAccountRealm(jwtUtils, userService)

	handler := config.NewRouter(&config.Handlers{
		Account: controller.NewAccountController(userService, jwtUtils),
		Blog:    controller.NewBlogController(blogService),
		User:    controller.NewUserController(userService),
		Filter:  shiro.NewJwtFilter(jwtUtils, realm),
	})

	return &App{
		Handler:     handler,
		JwtUtils:    jwtUtils,
		UserService: userService,
		BlogService: blogService,
		UserMapper:  um,
		BlogMapper:  bm,
	}
}

// Token mints a valid JWT for the given user id.
func (a *App) Token(userId int64) string {
	tok, err := a.JwtUtils.GenerateToken(userId)
	if err != nil {
		panic(err)
	}
	return tok
}
