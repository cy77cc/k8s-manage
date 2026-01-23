package users

import v1 "github.com/cy77cc/k8s-manage/api/users/v1"

type Users interface {
	Auth
	Roles
	Permissions
	// 创建用户
	CreateUser(v1.UserCreateReq) v1.UserResp
	// 获取用户
	GetUser() v1.UserResp
	// 更新用户
	UpdateUser(v1.UserUpdateReq) v1.UserResp
	// 删除用户
	DeleteUser()
	// 列出用户
	ListUsers(v1.UserListReq) v1.UserListResp
	// 分配角色给用户
	AssignRoleToUser(v1.AssignRoleBody)
	// 获取用户的角色列表
	ListUserRoles() v1.RoleListResp
	// 移除用户的角色
	RevokeRoleFromUser(v1.RevokeRoleBody)
}

type Auth interface {
	// 登录
	Login(v1.LoginReq) v1.TokenResp
	// 注册
	Register(v1.UserCreateReq) v1.TokenResp
	// 刷新token
	Refresh(v1.RefreshReq) v1.TokenResp
	// 登出
	Logout(v1.LogoutReq)
}

type Roles interface {
	// 创建角色
	CreateRole(v1.RoleCreateReq) v1.RoleResp
	// 获取角色信息
	GetRole() v1.RoleResp
	// 更新角色信息
	UpdateRole(v1.RoleUpdateReq) v1.RoleResp
	// 删除角色
	DeleteRole()
	// 获取角色列表
	ListRoles(v1.RoleListReq) v1.RoleListResp
	// 给角色分配权限
	GrantPermissionToRole(v1.GrantPermissionBody)
	// 获取角色的权限列表
	ListRolePermissions() v1.PermissionListResp
	// 移除角色的权限
	RevokePermissionFromRole(v1.RevokePermissionBody)
}

type Permissions interface {
	// 创建权限
	CreatePermission(v1.PermissionCreateReq) v1.PermissionResp
	// 获取权限信息
	GetPermission() v1.PermissionResp
	// 更新权限信息
	UpdatePermission(v1.PermissionUpdateReq) v1.PermissionResp
	// 删除权限
	DeletePermission()
	// 获取权限列表
	ListPermissions(v1.PermissionListReq) v1.PermissionListResp
}

type RBAC interface {
	// 检查用户权限
	CheckPermission(v1.CheckPermissionReq) v1.CheckPermissionResp
}