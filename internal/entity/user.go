package entity

import "github.com/markerhub/vueblog-go/internal/common/lang"

// User mirrors com.markerhub.entity.User (@TableName("m_user"), @Accessors(chain=true)).
//
// Avatar/Email/Password are pointers because the m_user columns are nullable
// and Jackson emits "email":null for a null field under ALWAYS inclusion.
type User struct {
	Id        *int64              `json:"id"`
	Username  string              `json:"username"`
	Avatar    *string             `json:"avatar"`
	Email     *string             `json:"email"`
	Password  *string             `json:"password"`
	Status    *int32              `json:"status"`
	Created   *lang.LocalDateTime `json:"created"`
	LastLogin *lang.LocalDateTime `json:"lastLogin"`
}

const UserTableName = "m_user"

func (u *User) SetId(id int64) *User       { u.Id = &id; return u }
func (u *User) SetUsername(v string) *User { u.Username = v; return u }
func (u *User) SetAvatar(v string) *User   { u.Avatar = &v; return u }
func (u *User) SetEmail(v string) *User    { u.Email = &v; return u }
func (u *User) SetPassword(v string) *User { u.Password = &v; return u }
func (u *User) SetStatus(v int32) *User    { u.Status = &v; return u }

// GetStatus mirrors user.getStatus(). AccountRealm compares it to -1; the Java
// code auto-unboxes an Integer, so a NULL status column throws NPE there.
func (u *User) GetStatus() *int32 { return u.Status }
