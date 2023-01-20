package handlers

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func GetRoles(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRolesRequest(r)
	if err != nil {
		Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	githubClient := github.NewGithub(Params(r).Token)

	findType, err := githubClient.FindType(request.Link)
	if err != nil {
		Log(r).WithError(err).Info("failed to get type")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if findType == "" {
		ape.Render(w, models.NewRolesResponse(false, "", ""))
		return
	}

	owned := ""
	if findType == data.Repository {
		owned, err = githubClient.FindRepoOwner(request.Link)
		if err != nil {
			Log(r).WithError(err).Infof("failed to get repo owner for ``")
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
	}

	ape.Render(w, models.NewRolesResponse(true, findType, owned))
}
