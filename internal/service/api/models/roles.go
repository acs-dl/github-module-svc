package models

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
)

var orgRepoRoles = []resources.Role{
	{Name: "Read", Value: "pull"},
	{Name: "Triage", Value: "triage"},
	{Name: "Write", Value: "push"},
	{Name: "Maintain", Value: "maintain"},
	{Name: "Admin", Value: "admin"},
}

var userRepoRoles = []resources.Role{
	{Name: "Write", Value: "push"},
}

var orgRoles = []resources.Role{
	{Name: "Member", Value: "member"},
	{Name: "Admin", Value: "admin"},
}

func NewRolesModel(roles []resources.Role) resources.Roles {
	result := resources.Roles{
		Key: resources.Key{
			ID:   "roles",
			Type: resources.ROLES,
		},
		Attributes: resources.RolesAttributes{
			Req:  true,
			List: roles,
		},
	}

	return result
}

func NewEmptyRolesModel() resources.Roles {
	result := resources.Roles{
		Key: resources.Key{
			ID:   "roles",
			Type: resources.ROLES,
		},
		Attributes: resources.RolesAttributes{
			Req:  false,
			List: nil,
		},
	}

	return result
}

func NewRolesResponse(found bool, typeTo, owned string) resources.RolesResponse {
	if !found {
		return resources.RolesResponse{
			Data: NewEmptyRolesModel(),
		}
	}

	if typeTo == data.Organization {
		return resources.RolesResponse{
			Data: NewRolesModel(orgRoles),
		}
	}

	if owned == data.UserOwned {
		return resources.RolesResponse{
			Data: NewRolesModel(userRepoRoles),
		}
	}

	return resources.RolesResponse{
		Data: NewRolesModel(orgRepoRoles),
	}
}
