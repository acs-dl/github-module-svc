package handlers

import (
	"fmt"
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId, err := requests.NewGetUserByIdRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	user, err := background.UsersQ(r).FilterById(&userId).Get()
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get user with id `%d`", userId)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if user == nil {
		background.Log(r).Errorf("no user with id `%d`", userId)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	permission, err := background.PermissionsQ(r).FilterByHasParent(false).FilterByParentLinks([]string{}...).FilterByGithubIds(user.GithubId).Get()
	if err != nil {
		background.Log(r).Errorf("failed to get submodule for user with id `%d`", userId)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	if permission != nil {
		accessLevel := fmt.Sprintf("%v", permission.AccessLevel)
		user.Submodule = &permission.Link
		user.AccessLevel = &accessLevel
	}

	ape.Render(w, models.NewUserResponse(*user))
}
