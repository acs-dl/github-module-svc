package handlers

import (
	"net/http"

	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/service/api/models"
	"github.com/acs-dl/github-module-svc/internal/service/api/requests"
	"github.com/acs-dl/github-module-svc/internal/service/background"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetRole(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRoleRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.AccessLevel == nil {
		background.Log(r).Errorf("no access level was provided")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	name := data.Roles[*request.AccessLevel]
	if name == "" {
		background.Log(r).Errorf("no such access level `%s`", *request.AccessLevel)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, models.NewRoleResponse(data.Roles[*request.AccessLevel], *request.AccessLevel))
}
