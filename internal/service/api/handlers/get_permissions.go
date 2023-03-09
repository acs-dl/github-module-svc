package handlers

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func GetPermissions(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetPermissionsRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	var userIds []int64
	if request.UserId != nil {
		userIds = append(userIds, *request.UserId)
	}

	var parentIds []int64

	statement := SubsQ(r).WithPermissions().FilterByUserIds(userIds...).
		FilterByHasParent(false).FilterByParentIds(parentIds...)

	totalCount := SubsQ(r).CountWithPermissions().FilterByUserIds(userIds...).
		FilterByHasParent(false).FilterByParentIds(parentIds...)

	if request.ParentLink != nil {
		permission, err := SubsQ(r).FilterByLinks(*request.ParentLink).Get()
		if err != nil {
			Log(r).WithError(err).Error("failed to get permission")
			ape.RenderErr(w, problems.InternalError())
			return
		}
		if permission == nil {
			ape.Render(w, models.NewUserPermissionListResponse([]data.Sub{}))
			return
		}

		SubsQ(r).ResetFilters()

		parentIds = append(parentIds, permission.Id)

		statement = SubsQ(r).WithPermissions().FilterByUserIds(userIds...).
			FilterByHasParent(false).FilterByParentIds(parentIds...)
		totalCount = SubsQ(r).CountWithPermissions().FilterByUserIds(userIds...).
			FilterByHasParent(false).FilterByParentIds(parentIds...)
	}

	var link = ""
	if request.Link != nil {
		link = *request.Link
		parentIds = nil

		statement = SubsQ(r).WithPermissions().FilterByUserIds(userIds...).
			SearchBy(link)
		totalCount = SubsQ(r).CountWithPermissions().FilterByUserIds(userIds...).
			SearchBy(link)
	}

	permissions, err := statement.Page(request.OffsetPageParams).Select()
	if err != nil {
		Log(r).WithError(err).Error("failed to get permissions")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	amount, err := totalCount.GetTotalCount()
	if err != nil {
		Log(r).WithError(err).Error("failed to get total count")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := models.NewUserPermissionListResponse(permissions)
	response.Meta.TotalCount = amount
	response.Links = data.GetOffsetLinksForPGParams(r, request.OffsetPageParams)

	ape.Render(w, response)
}
