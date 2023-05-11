package processor

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/acs-dl/github-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) sendUsers(uuid string, users []data.User) error {
	unverifiedUsers := make([]data.UnverifiedUser, 0)
	for i := range users {
		if users[i].Id != nil {
			continue
		}

		permission, err := p.permissionsQ.
			FilterByGithubIds(users[i].GithubId).
			FilterByGreaterTime(users[i].CreatedAt).
			FilterByHasParent(false).
			FilterByParentLinks([]string{}...).
			Get()
		if err != nil {
			p.log.WithError(err).Errorf("failed to select permissions by date `%s`", users[i].CreatedAt.String())
			return errors.Wrap(err, "failed to select permissions by date")
		}

		if permission == nil {
			continue
		}

		unverifiedUsers = append(unverifiedUsers, createUnverifiedUserFromModuleUser(users[i], permission.Link))
	}

	err := p.sender.SendMessageToCustomChannel("unverified-svc", p.buildUnverifiedUserListMessage(uuid, data.UnverifiedPayload{
		Action: SetUsersAction,
		Users:  unverifiedUsers,
	}))
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users to `unverified-svc`")
		return errors.Wrap(err, "failed to publish users to `unverified-svc`")
	}

	p.log.Infof("successfully published users to `unverified-svc`")
	return nil
}

func (p *processor) SendDeleteUser(uuid string, user data.User) error {
	unverifiedUsers := make([]data.UnverifiedUser, 0)

	unverifiedUsers = append(unverifiedUsers, createUnverifiedUserFromModuleUser(user, ""))

	err := p.sender.SendMessageToCustomChannel("unverified-svc", p.buildUnverifiedUserListMessage(uuid, data.UnverifiedPayload{
		Action: DeleteUsersAction,
		Users:  unverifiedUsers,
	}))
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users to `unverified-svc`")
		return errors.Wrap(err, "failed to publish users to `unverified-svc`")
	}

	p.log.Infof("successfully published users to `unverified-svc`")
	return nil
}

func (p *processor) buildUnverifiedUserListMessage(uuid string, unverifiedPayload data.UnverifiedPayload) *message.Message {
	marshaled, err := json.Marshal(unverifiedPayload)
	if err != nil {
		p.log.WithError(err).Errorf("failed to marshal response")
	}

	return &message.Message{
		UUID:     uuid,
		Metadata: nil,
		Payload:  marshaled,
	}
}

func createUnverifiedUserFromModuleUser(user data.User, submodule string) data.UnverifiedUser {
	return data.UnverifiedUser{
		CreatedAt: user.CreatedAt,
		Module:    data.ModuleName,
		Submodule: submodule,
		ModuleId:  fmt.Sprintf("%d", user.GithubId),
		Email:     nil,
		Name:      nil,
		Phone:     nil,
		Username:  &user.Username,
	}
}
