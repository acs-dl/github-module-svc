package handlers

import (
	"net/http"

	"github.com/pkg/errors"
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
		background.Log(r).Infof("no link was provided")
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	if request.Username == nil {
		background.Log(r).Infof("no username was provided")
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
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
			owned, err = getRepositoryOwnerType(pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperPQueue, githubClient, *request.Link)
			if err != nil {
				background.Log(r).WithError(err).Errorf("failed to get repository owner type")
				ape.RenderErr(w, problems.InternalError())
				return
			}
		}

		ape.Render(w, models.NewRolesResponse(true, permission.Type, owned, permission.AccessLevel))
		return
	}

	user, err := getUser(pqueue.PQueuesInstance(background.ParentContext(r.Context())).UsualPQueue, githubClient, *request.Username)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get user from api")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if user == nil {
		background.Log(r).Infof("no user with `%s` username in github", *request.Username)
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	typeSub, err := getLinkType(pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperPQueue, githubClient, *request.Link)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get link type")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if typeSub == nil {
		background.Log(r).Infof("nothing found with `%s` link in github", *request.Link)
		ape.Render(w, models.NewRolesResponse(false, "", "", ""))
		return
	}

	owned := data.OrganizationOwned
	if typeSub.Type == data.Repository {
		owned, err = getRepositoryOwnerType(pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperPQueue, githubClient, *request.Link)
		if err != nil {
			background.Log(r).WithError(err).Errorf("failed to get repository owner type")
			ape.RenderErr(w, problems.InternalError())
			return
		}
	}

	permission, err = getPermission(pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperPQueue, githubClient, *request.Link, *request.Username, typeSub.Type)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get permission for repository")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if permission == nil {
		ape.Render(w, models.NewRolesResponse(true, typeSub.Type, owned, ""))
		return
	}
	ape.Render(w, models.NewRolesResponse(true, typeSub.Type, owned, permission.AccessLevel))

}

func getUser(pq *pqueue.PriorityQueue, githubClient github.GithubClient, username string) (*data.User, error) {
	item, err := helpers.AddFunctionInPQueue(pq, any(githubClient.GetUserFromApi), []any{any(username)}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	user, ok := item.Response.Value.(*data.User)
	if !ok {
		return nil, errors.Errorf("wrong response type")
	}

	return user, nil
}

func getLinkType(pq *pqueue.PriorityQueue, githubClient github.GithubClient, link string) (*github.TypeSub, error) {
	item, err := helpers.AddFunctionInPQueue(pq, any(githubClient.FindType), []any{any(link)}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get type")
	}

	typeSub, ok := item.Response.Value.(*github.TypeSub)
	if !ok {
		return nil, errors.Errorf("wrong response type")
	}

	return typeSub, nil
}

func getRepositoryOwnerType(pq *pqueue.PriorityQueue, githubClient github.GithubClient, link string) (string, error) {
	item, err := helpers.AddFunctionInPQueue(pq, any(githubClient.FindRepositoryOwner), []any{any(link)}, pqueue.HighPriority)
	if err != nil {
		return "", errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return "", errors.Wrap(err, "some error while getting repo owner type")
	}
	var ok bool

	owned, ok := item.Response.Value.(string)
	if !ok {
		return "", errors.Errorf("wrong response type")
	}

	return owned, nil
}

func getPermission(pq *pqueue.PriorityQueue, githubClient github.GithubClient, link, username, typeTo string) (*data.Permission, error) {
	item, err := helpers.AddFunctionInPQueue(pq, any(githubClient.CheckUserFromApi), []any{any(link), any(username), any(typeTo)}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting repo owner type")
	}
	var ok bool

	owned, ok := item.Response.Value.(*github.CheckPermission)
	if !ok {
		return nil, errors.Errorf("wrong response type")
	}

	if owned == nil {
		return nil, nil
	}

	return &owned.Permission, nil
}
