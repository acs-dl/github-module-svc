package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetRoles(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRolesRequest(r)
	if err != nil {
		Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	githubClient := github.NewGithub(Params(r).Token)

	if request.Username != nil {
		permission, err := PermissionsQ(r).FilterByUsernames(*request.Username).FilterByLinks(*request.Link).Get()
		if err != nil {
			Log(r).WithError(err).Infof("failed to get permission from `%s` for `%s`", *request.Link, *request.Username)
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
		if permission != nil {
			owned := data.OrganizationOwned
			if permission.Type == data.Repository {
				owned, err = githubClient.FindRepoOwner(*request.Link)
				if err != nil {
					Log(r).WithError(err).Infof("failed to get repo owner for `%s`", *request.Link)
					ape.RenderErr(w, problems.BadRequest(err)...)
					return
				}
			}

			ape.Render(w, models.NewRolesResponse(true, permission.Type, owned, permission.AccessLevel))
			return
		}
	}

	findType, _, err := githubClient.FindType(*request.Link)
	if err != nil {
		Log(r).WithError(err).Info("failed to get type")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if findType == "" {
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	owned := data.OrganizationOwned
	if findType == data.Repository {
		owned, err = githubClient.FindRepoOwner(*request.Link)
		if err != nil {
			Log(r).WithError(err).Infof("failed to get repo owner for `%s`", *request.Link)
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
	}

	ape.Render(w, models.NewRolesResponse(true, findType, owned, ""))
}
