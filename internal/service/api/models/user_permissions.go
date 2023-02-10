package models

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
	"strconv"
)

func NewUserPermissionModel(permission data.Sub, counter int) resources.UserPermission {
	deployable := true
	if permission.Type == data.Repository {
		deployable = false
	}

	result := resources.UserPermission{
		Key: resources.Key{
			ID:   strconv.Itoa(counter),
			Type: resources.USER_PERMISSION,
		},
		Attributes: resources.UserPermissionAttributes{
			Username: permission.Username,
			ModuleId: permission.GithubId,
			Path:     permission.Path,
			UserId:   permission.UserId,
			Level:    permission.Nlevel,
			Type:     permission.Type,
			Link:     permission.Link,
			AccessLevel: resources.AccessLevel{
				Name:  data.Roles[permission.AccessLevel],
				Value: permission.AccessLevel,
			},
			Deployable: deployable,
			ExpiresAt:  permission.ExpiresAt,
		},
	}

	return result
}

func NewUserPermissionList(permissions []data.Sub) []resources.UserPermission {
	result := make([]resources.UserPermission, len(permissions))
	for i, permission := range permissions {
		result[i] = NewUserPermissionModel(permission, i)
	}
	return result
}

func NewUserPermissionListResponse(permissions []data.Sub) UserPermissionListResponse {
	return UserPermissionListResponse{
		Data: NewUserPermissionList(permissions),
	}
}

type UserPermissionListResponse struct {
	Meta  Meta                       `json:"meta"`
	Data  []resources.UserPermission `json:"data"`
	Links *resources.Links           `json:"links"`
}
