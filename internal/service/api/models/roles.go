package models

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
)

var orgRepoRoles = []resources.Role{
	{Name: "Read", Value: "read"},
	{Name: "Triage", Value: "triage"},
	{Name: "Write", Value: "write"},
	{Name: "Maintain", Value: "maintain"},
	{Name: "Admin", Value: "admin"},
}

var userRepoRoles = []resources.Role{
	{Name: "Write", Value: "write"},
}

var orgRoles = []resources.Role{
	{Name: "Member", Value: "member"},
	{Name: "Admin", Value: "admin"},
}

func NewRolesModel(found bool, roles []resources.Role) resources.Roles {
	result := resources.Roles{
		Key: resources.Key{
			ID:   "roles",
			Type: resources.ROLES,
		},
		Attributes: resources.RolesAttributes{
			Req:  found,
			List: roles,
		},
	}

	return result
}

func NewRolesResponse(found bool, typeTo, owned, current string) resources.RolesResponse {
	if !found {
		return resources.RolesResponse{
			Data: NewRolesModel(found, []resources.Role{}),
		}
	}

	if typeTo == data.Organization {
		return resources.RolesResponse{
			Data: NewRolesModel(found, newRolesArray(current, orgRoles)),
		}
	}

	if owned == data.UserOwned {
		return resources.RolesResponse{
			Data: NewRolesModel(found, newRolesArray(current, userRepoRoles)),
		}
	}

	return resources.RolesResponse{
		Data: NewRolesModel(found, newRolesArray(current, orgRepoRoles)),
	}
}

func newRolesArray(current string, roles []resources.Role) []resources.Role {
	var result []resources.Role

	for _, role := range roles {
		if role.Value != current {
			result = append(result, role)
		}
	}

	return result
}
