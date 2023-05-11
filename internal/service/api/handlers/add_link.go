package handlers

import (
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func AddLink(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewAddLinkRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Error("failed to parse add link request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	err = background.LinksQ(r).Insert(data.Link{Link: request.Data.Attributes.Link})
	if err != nil {
		background.Log(r).WithError(err).Error("failed to save new link")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	background.Log(r).Infof("successfully created link `%s`", request.Data.Attributes.Link)
	w.WriteHeader(http.StatusAccepted)
	ape.Render(w, http.StatusAccepted)
}
