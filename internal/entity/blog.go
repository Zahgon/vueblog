// Package entity mirrors com.markerhub.entity.
package entity

import "github.com/markerhub/vueblog-go/internal/common/lang"

// Blog mirrors com.markerhub.entity.Blog (@TableName("m_blog"), @Accessors(chain=true)).
//
// Field order matches the Java declaration order so Jackson's output order is
// reproduced. Id/UserId/Status are pointers because the Java fields are boxed
// (Long/Integer) and the edit endpoint branches on blog.getId() != null.
type Blog struct {
	Id          *int64          `json:"id"`
	UserId      *int64          `json:"userId"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Content     string          `json:"content"`
	Created     *lang.LocalDate `json:"created"` // @JsonFormat(pattern="yyyy-MM-dd")
	Status      *int32          `json:"status"`
}

// TableName mirrors @TableName("m_blog").
const BlogTableName = "m_blog"

// Chained setters mirror @Accessors(chain = true).

func (b *Blog) SetId(id int64) *Blog          { b.Id = &id; return b }
func (b *Blog) SetUserId(userId int64) *Blog  { b.UserId = &userId; return b }
func (b *Blog) SetTitle(v string) *Blog       { b.Title = v; return b }
func (b *Blog) SetDescription(v string) *Blog { b.Description = v; return b }
func (b *Blog) SetContent(v string) *Blog     { b.Content = v; return b }
func (b *Blog) SetCreated(v *lang.LocalDate) *Blog {
	b.Created = v
	return b
}
func (b *Blog) SetStatus(v int32) *Blog { b.Status = &v; return b }

// GetId mirrors the Lombok getter used by BlogController.edit.
func (b *Blog) GetId() *int64 { return b.Id }

// GetUserId mirrors blog.getUserId(). Callers that dereference reproduce the
// Java NullPointerException path - see controller.BlogController.Edit.
func (b *Blog) GetUserId() *int64 { return b.UserId }
