package requests

import (
	"net/http"

	"gitlab.com/distributed_lab/urlval"
)

type GetRolesRequest struct {
	Link     *string `filter:"link"`
	Username *string `filter:"username"`
}

func NewGetRolesRequest(r *http.Request) (GetRolesRequest, error) {
	var request GetRolesRequest

	err := urlval.Decode(r.URL.Query(), &request)

	return request, err
}
