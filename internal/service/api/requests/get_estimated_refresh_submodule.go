package requests

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/urlval"
)

type GetEstimatedRefreshSubmoduleRequest struct {
	Submodules *string `filter:"submodule"`
}

func NewGetEstimatedRefreshSubmoduleRequest(r *http.Request) (GetEstimatedRefreshSubmoduleRequest, error) {
	var request GetEstimatedRefreshSubmoduleRequest

	err := urlval.Decode(r.URL.Query(), &request)
	if err != nil {
		return request, nil
	}

	return request, request.validate()
}

func (r *GetEstimatedRefreshSubmoduleRequest) validate() error {
	return validation.Errors{
		"submodules": validation.Validate(&r.Submodules, validation.Required),
	}.Filter()
}
