package models

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
)

func NewUserPermissionModel(permission data.Permission) resources.UserPermission {
	result := resources.UserPermission{
		Key: resources.Key{
			ID:   permission.RequestId,
			Type: resources.USER_PERMISSION,
		},
		Attributes: resources.UserPermissionAttributes{
			Username:   permission.Username,
			GithubId:   permission.GithubId,
			Link:       permission.Link,
			Permission: permission.Permission,
			Type:       permission.Type,
		},
	}

	return result
}

func NewUserPermissionList(permissions []data.Permission) []resources.UserPermission {
	result := make([]resources.UserPermission, len(permissions))
	for i, permission := range permissions {
		result[i] = NewUserPermissionModel(permission)
	}
	return result
}

func NewUserPermissionListResponse(permissions []data.Permission) resources.UserPermissionListResponse {
	return resources.UserPermissionListResponse{
		Data: NewUserPermissionList(permissions),
	}
}
