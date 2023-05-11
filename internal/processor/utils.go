package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) getLinkType(link string, priority int) (string, error) {
	checkType, err := github.GetPermissionWithType(p.pqueues.SuperUserPQueue, any(p.githubClient.FindType), []any{any(link)}, priority)
	if err != nil {
		return "", errors.Wrap(err, "failed to get link type")
	}

	if checkType == nil {
		return "", errors.New("no type was found ")
	}

	if validation.Validate(checkType.Type, validation.In(data.Organization, data.Repository)) != nil {
		return "", errors.Wrap(err, "something wrong with link type")
	}

	return checkType.Type, nil
}

func (p *processor) isUserInSubmodule(link, username, typeTo string) (bool, error) {
	permission, err := github.GetPermission(
		p.pqueues.SuperUserPQueue,
		any(p.githubClient.CheckUserFromApi),
		[]any{any(link), any(username), any(typeTo)},
		pqueue.NormalPriority,
	)
	if err != nil {
		return false, errors.Wrap(err, "some error while checking link type api")
	}

	return permission != nil, err
}
