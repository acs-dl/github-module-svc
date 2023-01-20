package requests

import "net/http"

func NewGetUserPermissionsRequest(r *http.Request) (int64, error) {
	return RetrieveId(r)
}
