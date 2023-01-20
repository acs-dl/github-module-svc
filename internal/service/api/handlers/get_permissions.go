package handlers

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func GetPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := PermissionsQ(r).Select()
	if err != nil {
		Log(r).WithError(err).Error("failed to get permissions")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.NewPermissionListResponse(permissions))
}
