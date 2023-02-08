package models

import (
	"gitlab.com/distributed_lab/acs/github-module/resources"
)

func NewInputsModel() resources.Inputs {
	result := resources.Inputs{
		Key: resources.Key{
			ID:   "inputs",
			Type: resources.INPUTS,
		},
		Attributes: resources.InputsAttributes{
			Username:    "string",
			Link:        "string",
			AccessLevel: "string",
		},
	}

	return result
}

func NewInputsResponse() resources.InputsResponse {
	return resources.InputsResponse{
		Data: NewInputsModel(),
	}
}
