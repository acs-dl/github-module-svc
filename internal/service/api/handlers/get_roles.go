package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetRoles(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRolesRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	githubClient := github.NewGithub(background.Config(r.Context()))

	if request.Username != nil {
		permission, err := background.PermissionsQ(r).FilterByUsernames(*request.Username).FilterByLinks(*request.Link).Get()
		if err != nil {
			background.Log(r).WithError(err).Infof("failed to get permission from `%s` for `%s`", *request.Link, *request.Username)
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
		if permission != nil {
			owned := data.OrganizationOwned
			if permission.Type == data.Repository {
				item, err := helpers.AddFunctionInPqueue(pqueue.PQueue(r.Context()), any(githubClient.FindRepoOwner), []any{any(*request.Link)}, pqueue.HighPriority)
				if err != nil {
					background.Log(r).WithError(err).Errorf("failed to add function in pqueue")
					ape.RenderErr(w, problems.InternalError())
					return
				}

				err = item.Response.Error
				if err != nil {
					background.Log(r).WithError(err).Errorf("some error while getting repo owner type")
					ape.RenderErr(w, problems.InternalError())
					return
				}
				var ok bool
				owned, ok = item.Response.Value.(string)
				if !ok {
					background.Log(r).Errorf("wrong response type")
					ape.RenderErr(w, problems.InternalError())
					return
				}
			}

			ape.Render(w, models.NewRolesResponse(true, permission.Type, owned, permission.AccessLevel))
			return
		}
	}

	item, err := helpers.AddFunctionInPqueue(pqueue.PQueue(r.Context()), any(githubClient.FindType), []any{any(*request.Link)}, pqueue.HighPriority)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to add function in pqueue")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = item.Response.Error
	if err != nil {
		background.Log(r).WithError(err).Info("failed to get type")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}
	typeSub, ok := item.Response.Value.(*github.TypeSub)
	if !ok {
		background.Log(r).Errorf("wrong response type")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if typeSub == nil {
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	owned := data.OrganizationOwned
	if typeSub.Type == data.Repository {
		item, err = helpers.AddFunctionInPqueue(pqueue.PQueue(r.Context()), any(githubClient.FindRepoOwner), []any{any(*request.Link)}, pqueue.HighPriority)
		if err != nil {
			background.Log(r).WithError(err).Errorf("failed to add function in pqueue")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		err = item.Response.Error
		if err != nil {
			background.Log(r).WithError(err).Infof("failed to get repo owner for `%s`", *request.Link)
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}
		owned, ok = item.Response.Value.(string)
		if !ok {
			background.Log(r).Errorf("wrong response type")
			ape.RenderErr(w, problems.InternalError())
			return
		}
	}

	ape.Render(w, models.NewRolesResponse(true, typeSub.Type, owned, ""))
}
