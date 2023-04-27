package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CheckSubmodule(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCheckSubmoduleRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		background.Log(r).Errorf("no link was provided")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	sub, err := background.SubsQ(r).FilterByLinks(*request.Link).Get()
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get link `%s`", *request.Link)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if sub != nil {
		ape.Render(w, models.NewLinkResponse(sub.Path, true))
		return
	}

	githubClient := github.GithubClientInstance(background.ParentContext(r.Context()))

	typeSub, err := getLinkType(pqueue.PQueuesInstance(background.ParentContext(r.Context())).SuperUserPQueue, githubClient, *request.Link)
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get type from api")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if typeSub != nil {
		ape.Render(w, models.NewLinkResponse(typeSub.Sub.Path, true))
		return
	}

	background.Log(r).Warnf("no group/project was found")
	ape.Render(w, models.NewLinkResponse("", false))
}
