package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/resources"
	"gitlab.com/distributed_lab/ape"
)

func GetRolesMap(w http.ResponseWriter, r *http.Request) {
	result := newModuleRolesResponse()

	for key, val := range data.Roles {
		result.Data.Attributes[key] = val
	}

	ape.Render(w, result)
}

func newModuleRolesResponse() ModuleRolesResponse {
	return ModuleRolesResponse{
		Data: ModuleRoles{
			Key: resources.Key{
				ID:   "0",
				Type: resources.MODULES,
			},
			Attributes: Roles{},
		},
	}
}

type ModuleRolesResponse struct {
	Data ModuleRoles `json:"data"`
}

type Roles map[string]string
type ModuleRoles struct {
	resources.Key
	Attributes Roles `json:"attributes"`
}
