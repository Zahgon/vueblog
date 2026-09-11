package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/markerhub/vueblog-go/internal/common/exception"
	"github.com/markerhub/vueblog-go/internal/common/lang"
	"github.com/markerhub/vueblog-go/internal/entity"
	"github.com/markerhub/vueblog-go/internal/mybatisplus"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/shiro"
	"github.com/markerhub/vueblog-go/internal/validation"
)

// BlogController mirrors com.markerhub.controller.BlogController.
type BlogController struct{ BlogService service.BlogService }

func NewBlogController(bs service.BlogService) *BlogController {
	return &BlogController{BlogService: bs}
}

// List mirrors @GetMapping("/blogs") with
// @RequestParam(defaultValue = "1") Integer currentPage.
//
// Spring converts the query parameter with Integer.parseInt; a non-numeric
// value raises MethodArgumentTypeMismatchException. That is not handled
// explicitly, so it falls through to the RuntimeException handler -> 400.
func (c *BlogController) List(w http.ResponseWriter, r *http.Request) *lang.Result {
	currentPage := int64(1)
	if raw := r.URL.Query().Get("currentPage"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			panic(&exception.RuntimeException{
				Message: `Failed to convert value of type 'java.lang.String' to required type 'java.lang.Integer'; nested exception is java.lang.NumberFormatException: For input string: "` + raw + `"`,
			})
		}
		currentPage = v
	}

	page := mybatisplus.NewPage[*entity.Blog](currentPage, 5)
	pageData, err := c.BlogService.Page(page, mybatisplus.NewQueryWrapper().OrderByDesc("created"))
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}

	return lang.Succ(pageData)
}

// Detail mirrors @GetMapping("/blog/{id}") with @PathVariable Long id.
func (c *BlogController) Detail(w http.ResponseWriter, r *http.Request, idRaw string) *lang.Result {
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		panic(&exception.RuntimeException{
			Message: `Failed to convert value of type 'java.lang.String' to required type 'java.lang.Long'; nested exception is java.lang.NumberFormatException: For input string: "` + idRaw + `"`,
		})
	}

	blog, err := c.BlogService.GetById(id)
	if err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	exception.NotNull(blog, "该博客已被删除")

	return lang.Succ(blog)
}

// Edit mirrors @RequiresAuthentication @PostMapping("/blog/edit").
//
// Two faithful details worth calling out:
//
//  1. When an id is supplied for a row that does not exist, blogService.getById
//     returns null and temp.getUserId() throws NullPointerException. The Java
//     code has no guard, so this returns 400 with "msg":null rather than a
//     friendly message. Preserved deliberately.
//  2. BeanUtil.copyProperties(blog, temp, "id","userId","created","status")
//     copies every property EXCEPT those four, so only title, description and
//     content move across. A client cannot rewrite ownership or timestamps.
func (c *BlogController) Edit(w http.ResponseWriter, r *http.Request) *lang.Result {
	RequiresAuthentication(r)

	var blog entity.Blog
	if err := json.NewDecoder(r.Body).Decode(&blog); err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}
	validation.ValidateBlog(&blog)

	var temp *entity.Blog
	if blog.GetId() != nil {
		got, err := c.BlogService.GetById(*blog.GetId())
		if err != nil {
			panic(&exception.RuntimeException{Message: err.Error()})
		}
		temp = got

		profile := shiro.GetProfile(r.Context())
		// System.out.println(ShiroUtil.getProfile().getId()) - the original
		// prints the id here; an anonymous subject would NPE, but
		// @RequiresAuthentication has already guaranteed a principal.
		log.Println(derefId(profile.GetId()))

		// temp.getUserId() on a null temp is the NPE described above.
		if temp == nil {
			exception.ThrowNPE()
		}
		if temp.GetUserId() == nil {
			exception.ThrowNPE()
		}
		exception.IsTrue(*temp.GetUserId() == *profile.GetId(), "没有权限编辑")
	} else {
		profile := shiro.GetProfile(r.Context())
		temp = &entity.Blog{}
		temp.SetUserId(*profile.GetId())
		temp.SetCreated(lang.NewLocalDate(time.Now()))
		temp.SetStatus(0)
	}

	copyPropertiesIgnoring(&blog, temp)

	if _, err := c.BlogService.SaveOrUpdate(temp); err != nil {
		panic(&exception.RuntimeException{Message: err.Error()})
	}

	return lang.Succ(nil)
}

// copyPropertiesIgnoring mirrors
// BeanUtil.copyProperties(blog, temp, "id", "userId", "created", "status").
func copyPropertiesIgnoring(src, dst *entity.Blog) {
	dst.Title = src.Title
	dst.Description = src.Description
	dst.Content = src.Content
}

func derefId(id *int64) any {
	if id == nil {
		return nil
	}
	return *id
}
