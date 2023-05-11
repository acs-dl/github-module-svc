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

func GetUnverifiedUsers(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetUnverifiedUsersRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	username := ""
	if request.Username != nil {
		username = *request.Username
	}

	totalCount, err := background.UsersQ(r).Count().FilterById(nil).SearchBy(username).GetTotalCount()
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to select to get total count from db")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	users, err := background.UsersQ(r).SearchBy(username).FilterById(nil).Page(request.OffsetPageParams).Select()
	if err != nil {
		background.Log(r).WithError(err).Errorf("failed to select unverified users from db")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := models.NewUserListResponse(users, request.OffsetPageParams.PageNumber*request.OffsetPageParams.Limit)
	response.Meta.TotalCount = totalCount
	response.Links = data.GetOffsetLinksForPGParams(r, request.OffsetPageParams)
	ape.Render(w, response)
}
