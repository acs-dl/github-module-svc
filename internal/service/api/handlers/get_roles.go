package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/acs/github-module/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func GetRoles(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRolesRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		background.Log(r).Infof("no link was provided")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	if request.Username == nil {
		background.Log(r).Infof("no username was provided")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	githubClient := github.GithubClientInstance(background.ParentContext(r.Context()))

	permission, err := background.PermissionsQ(r).FilterByUsernames(*request.Username).FilterByLinks(*request.Link).Get()
	if err != nil {
		background.Log(r).WithError(err).Infof("failed to get permission from `%s` for `%s`", *request.Link, *request.Username)
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if permission != nil {
		owned := data.OrganizationOwned
		if permission.Type == data.Repository {
			owned, err = github.GetString(
				pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperUserPQueue,
				any(githubClient.FindRepositoryOwner),
				[]any{any(*request.Link)},
				pqueue.HighPriority,
			)
			if err != nil {
				background.Log(r).WithError(err).Errorf("failed to get repository owner type")
				ape.RenderErr(w, problems.InternalError())
				return
			}
		}

		ape.Render(w, models.NewRolesResponse(true, permission.Type, owned, permission.AccessLevel))
		return
	}

	response, err := checkRemoteUser(r, *request.Username, *request.Link)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to check remote user")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if response == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, response)

}

func checkRemoteUser(r *http.Request, username, link string) (*resources.RolesResponse, error) {
	githubClient := github.GithubClientInstance(background.ParentContext(r.Context()))

	user, err := github.GetUser(
		pqueue.PQueuesInstance(background.ParentContext(r.Context())).UserPQueue,
		any(githubClient.GetUserFromApi),
		[]any{any(username)},
		pqueue.HighPriority,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from api")
	}

	if user == nil {
		background.Log(r).Warnf("no user `%s` in github", username)
		return nil, nil
	}

	typeSub, err := github.GetPermissionWithType(
		pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperUserPQueue,
		any(githubClient.FindType),
		[]any{any(link)},
		pqueue.HighPriority,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link type")
	}

	if typeSub == nil {
		background.Log(r).Warnf("nothing found for `%s` in github", username)
		return nil, nil
	}

	owned := data.OrganizationOwned
	if typeSub.Type == data.Repository {
		owned, err = github.GetString(
			pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperUserPQueue,
			any(githubClient.FindRepositoryOwner),
			[]any{any(link)},
			pqueue.HighPriority,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get repository owner")
		}
	}

	permission, err := github.GetPermission(
		pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperUserPQueue,
		any(githubClient.CheckUserFromApi),
		[]any{any(link), any(username), any(typeSub.Type)},
		pqueue.HighPriority,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check user from api")
	}

	if permission != nil {
		return nil, nil
	}

	response := models.NewRolesResponse(true, typeSub.Type, owned, "")
	return &response, nil
}
