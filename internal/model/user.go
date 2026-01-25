package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username      string `gorm:"column:username;unique" json:"username"`
	PasswordHash  string `gorm:"column:password_hash" json:"-"`
	Email         string `gorm:"column:email" json:"email"`
	Phone         string `gorm:"column:phone" json:"phone"`
	Avatar        string `gorm:"column:avatar" json:"avatar"`
	Status        int8   `gorm:"column:status" json:"status"`
	CreateTime    int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime    int64  `gorm:"column:update_time" json:"update_time"`
	LastLoginTime int64  `gorm:"column:last_login_time" json:"last_login_time"`
}

func (User) TableName() string {
	return "users"
}

type UserRole struct {
	gorm.Model
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID int64 `gorm:"column:user_id" json:"user_id"`
	RoleID int64 `gorm:"column:role_id" json:"role_id"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type Role struct {
	gorm.Model
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code;unique" json:"code"`
	Description string `gorm:"column:description" json:"description"`
	Status      int8   `gorm:"column:status" json:"status"`
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
}

func (Role) TableName() string {
	return "roles"
}

type RolePermission struct {
	gorm.Model
	ID           int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID       int64 `gorm:"column:role_id" json:"role_id"`
	PermissionID int64 `gorm:"column:permission_id" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

type Permission struct {
	gorm.Model
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code;unique" json:"code"`
	Type        int8   `gorm:"column:type" json:"type"`
	Resource    string `gorm:"column:resource" json:"resource"`
	Action      string `gorm:"column:action" json:"action"`
	Description string `gorm:"column:description" json:"description"`
	Status      int8   `gorm:"column:status" json:"status"`
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
}

func (Permission) TableName() string {
	return "permissions"
}

type AuthRefreshToken struct {
	gorm.Model
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64  `gorm:"column:user_id" json:"user_id"`
	Token      string `gorm:"column:token" json:"token"`
	Expires    int64  `gorm:"column:expires" json:"expires"`
	Revoked    int8   `gorm:"column:revoked" json:"revoked"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
}

func (AuthRefreshToken) TableName() string {
	return "auth_refresh_tokens"
}
