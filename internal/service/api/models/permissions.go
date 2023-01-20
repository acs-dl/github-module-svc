package models

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
)

func NewPermissionModel(permission data.Permission) resources.Permission {
	var userId int64 = 0
	if permission.UserId != nil {
		userId = *permission.UserId
	}
	result := resources.Permission{
		Key: resources.Key{
			ID:   permission.RequestId,
			Type: resources.PERMISSION,
		},
		Attributes: resources.PermissionAttributes{
			UserId:     userId,
			Username:   permission.Username,
			GithubId:   permission.GithubId,
			Link:       permission.Link,
			Permission: permission.Permission,
			Type:       permission.Type,
		},
	}

	return result
}

func NewPermissionList(permissions []data.Permission) []resources.Permission {
	result := make([]resources.Permission, len(permissions))
	for i, permission := range permissions {
		result[i] = NewPermissionModel(permission)
	}
	return result
}

func NewPermissionListResponse(permissions []data.Permission) resources.PermissionListResponse {
	return resources.PermissionListResponse{
		Data: NewPermissionList(permissions),
	}
}
