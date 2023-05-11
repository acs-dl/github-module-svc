package handlers

import (
	"net/http"

	"github.com/acs-dl/github-module-svc/internal/service/api/requests"
	"github.com/acs-dl/github-module-svc/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func RemoveLink(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewRemoveLinkRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Error("failed to parse remove link request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	err = background.LinksQ(r).FilterByLinks(request.Data.Attributes.Link).Delete()
	if err != nil {
		background.Log(r).WithError(err).Error("failed to delete link")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	background.Log(r).Infof("successfully removed link `%s`", request.Data.Attributes.Link)
	w.WriteHeader(http.StatusAccepted)
	ape.Render(w, http.StatusAccepted)
}
