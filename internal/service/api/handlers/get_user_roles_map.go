package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/ape"
)

func GetUserRolesMap(w http.ResponseWriter, r *http.Request) {
	result := newModuleRolesResponse()

	result.Data.Attributes["super_admin"] = data.Roles["admin"]
	result.Data.Attributes["admin"] = data.Roles["member"]
	result.Data.Attributes["user"] = data.Roles["read"]

	ape.Render(w, result)
}
