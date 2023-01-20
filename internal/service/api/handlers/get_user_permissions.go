package handlers

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userId, err := requests.NewGetUserPermissionsRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	user, err := UsersQ(r).GetById(userId)
	if err != nil {
		Log(r).WithError(err).Error("failed to get user")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if user == nil {
		Log(r).Error("no user with such id")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	permissions, err := PermissionsQ(r).JoinsModule().FilterByUserIds(*user.Id).Select()
	if err != nil {
		Log(r).WithError(err).Error("failed to get permissions")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.NewUserPermissionListResponse(permissions))
}
