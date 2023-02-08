package requests

import (
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/urlval"
	"net/http"
)

type GetPermissionsRequest struct {
	pgdb.OffsetPageParams

	Link   *string `filter:"link"`
	UserId *int64  `filter:"userId"`
}

func NewGetPermissionsRequest(r *http.Request) (GetPermissionsRequest, error) {
	var request GetPermissionsRequest

	err := urlval.Decode(r.URL.Query(), &request)

	return request, err
}
