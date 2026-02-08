package model

import (
	"github.com/gookit/validate"
	"gorm.io/gorm"
)

type UserID int64

// User 用户表
type User struct {
	ID            UserID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username      string `gorm:"column:username;type:varchar(64);not null;unique" json:"username" validate:"required|minLen:7"`
	PasswordHash  string `gorm:"column:password_hash;type:varchar(255);not null" json:"password_hash"`
	Email         string `gorm:"column:email;type:varchar(128);not null;default:''" json:"email" validate:"email"`
	Phone         string `gorm:"column:phone;type:varchar(32);not null;default:''" json:"phone" validate:"phone"`
	Avatar        string `gorm:"column:avatar;type:varchar(255);not null;default:''" json:"avatar"`
	Status        int8   `gorm:"column:status;not null;default:1" json:"status"`
	CreateTime    int64  `gorm:"column:create_time;not null;default:0;autoCreateTime" json:"create_time"`
	UpdateTime    int64  `gorm:"column:update_time;not null;default:0;autoUpdateTime" json:"update_time"`
	LastLoginTime int64  `gorm:"column:last_login_time;not null;default:0" json:"last_login_time"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	v := validate.Struct(u)
	if !v.Validate() {
		return v.Errors
	}
	return nil
}

// UserRole 用户角色关联表
type UserRole struct {
	ID     UserID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID int64  `gorm:"column:user_id;not null" json:"user_id"`
	RoleID int64  `gorm:"column:role_id;not null" json:"role_id"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// Role 角色表
type Role struct {
	ID          UserID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);not null;default:''" json:"name"`
	Code        string `gorm:"column:code;type:varchar(64);not null;unique;default:''" json:"code"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Status      int8   `gorm:"column:status;not null;default:1" json:"status"`
	CreateTime  int64  `gorm:"column:create_time;not null;default:0;autoCreateTime" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time;not null;default:0;autoUpdateTime" json:"update_time"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           UserID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID       int64  `gorm:"column:role_id;not null" json:"role_id"`
	PermissionID int64  `gorm:"column:permission_id;not null" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// Permission 权限表
type Permission struct {
	ID          UserID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);not null;default:''" json:"name"`
	Code        string `gorm:"column:code;type:varchar(128);not null;unique;default:''" json:"code"`
	Type        int8   `gorm:"column:type;not null;default:0" json:"type"`
	Resource    string `gorm:"column:resource;type:varchar(255);not null;default:''" json:"resource"`
	Action      string `gorm:"column:action;type:varchar(32);not null;default:''" json:"action"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Status      int8   `gorm:"column:status;not null;default:1" json:"status"`
	CreateTime  int64  `gorm:"column:create_time;not null;default:0;autoCreateTime" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time;not null;default:0;autoUpdateTime" json:"update_time"`
}

func (Permission) TableName() string {
	return "permissions"
}
