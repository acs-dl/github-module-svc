package handlers

import (
	"net/http"
	"strings"

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

	link := strings.ToLower(*request.Link)

	sub, err := background.SubsQ(r).FilterByLinks(link).Get()
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to get link `%s`", link)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if sub != nil {
		ape.Render(w, models.NewLinkResponse(sub.Path, true))
		return
	}

	background.Log(r).Warnf("no group/project was found")
	ape.Render(w, models.NewLinkResponse("", false))
}
