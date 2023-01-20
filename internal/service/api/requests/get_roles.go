package requests

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
)

type GetRolesRequest struct {
	Module string `json:"module"`
	Link   string `json:"data"`
}

func NewGetRolesRequest(r *http.Request) (GetRolesRequest, error) {
	var request GetRolesRequest

	request.Module = r.URL.Query().Get("module")
	request.Link = r.URL.Query().Get("link")

	return request, request.validate()
}

func (r *GetRolesRequest) validate() error {
	return validation.Errors{
		"module": validation.Validate(&r.Module, validation.Required),
		"link":   validation.Validate(&r.Link, validation.Required),
	}.Filter()
}
