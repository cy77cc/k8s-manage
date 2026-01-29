package model

import "gorm.io/gorm"

// User 用户表
type User struct {
	gorm.Model
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`  // 主键ID
	Username      string `gorm:"column:username;unique" json:"username"`        // 用户名
	PasswordHash  string `gorm:"column:password_hash" json:"password_hash"`     // 密码哈希
	Email         string `gorm:"column:email" json:"email"`                     // 邮箱
	Phone         string `gorm:"column:phone" json:"phone"`                     // 手机号
	Avatar        string `gorm:"column:avatar" json:"avatar"`                   // 头像
	Status        int8   `gorm:"column:status" json:"status"`                   // 状态 1:正常 2:禁用
	CreateTime    int64  `gorm:"column:create_time" json:"create_time"`         // 创建时间
	UpdateTime    int64  `gorm:"column:update_time" json:"update_time"`         // 更新时间
	LastLoginTime int64  `gorm:"column:last_login_time" json:"last_login_time"` // 最后登录时间
}

func (User) TableName() string {
	return "users"
}

// UserRole 用户角色关联表
type UserRole struct {
	gorm.Model
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // 主键ID
	UserID int64 `gorm:"column:user_id" json:"user_id"`                // 用户ID
	RoleID int64 `gorm:"column:role_id" json:"role_id"`                // 角色ID
}

func (UserRole) TableName() string {
	return "user_roles"
}

// Role 角色表
type Role struct {
	gorm.Model
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // 主键ID
	Name        string `gorm:"column:name" json:"name"`                      // 角色名称
	Code        string `gorm:"column:code;unique" json:"code"`               // 角色唯一标识
	Description string `gorm:"column:description" json:"description"`        // 描述
	Status      int8   `gorm:"column:status" json:"status"`                  // 状态
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`        // 创建时间
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`        // 更新时间
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色权限关联表
type RolePermission struct {
	gorm.Model
	ID           int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // 主键ID
	RoleID       int64 `gorm:"column:role_id" json:"role_id"`                // 角色ID
	PermissionID int64 `gorm:"column:permission_id" json:"permission_id"`    // 权限ID
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// Permission 权限表
type Permission struct {
	gorm.Model
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // 主键ID
	Name        string `gorm:"column:name" json:"name"`                      // 权限名称
	Code        string `gorm:"column:code;unique" json:"code"`               // 权限标识
	Type        int8   `gorm:"column:type" json:"type"`                      // 类型 1:菜单 2:按钮 3:API
	Resource    string `gorm:"column:resource" json:"resource"`              // 资源路径
	Action      string `gorm:"column:action" json:"action"`                  // 请求方法
	Description string `gorm:"column:description" json:"description"`        // 描述
	Status      int8   `gorm:"column:status" json:"status"`                  // 状态
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`        // 创建时间
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`        // 更新时间
}

func (Permission) TableName() string {
	return "permissions"
}
