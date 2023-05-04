package processor

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateGetUsers(msg data.ModulePayload) error {
	return validation.Errors{
		"link": validation.Validate(msg.Link, validation.Required),
	}.Filter()
}

func (p *processor) HandleGetUsersAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateGetUsers(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	msg.Type, err = p.getLinkType(msg.Link, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get link type from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting link type api")
	}

	permissions, err := github.GetPermissions(p.pqueues.SuperUserPQueue, any(p.githubClient.GetUsersFromApi), []any{any(msg.Link), any(msg.Type)}, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get users from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting users from api")
	}

	usersToUnverified := make([]data.User, 0)

	for _, permission := range permissions {
		//api doesn't return role for organization members
		if msg.Type == data.Organization {
			checkPermission, err := github.GetPermission(
				p.pqueues.SuperUserPQueue,
				any(p.githubClient.CheckOrganizationCollaborator), []any{any(msg.Link), any(permission.Username)},
				pqueue.LowPriority)
			if err != nil {
				p.log.WithError(err).Errorf("failed to get permission from api for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get permission from api")
			}
			if checkPermission == nil {
				p.log.Errorf("user is not in organization for message action with id `%s`", msg.RequestId)
				return errors.Errorf("user is not in organization")
			}

			permission.AccessLevel = checkPermission.AccessLevel
		}

		err = p.managerQ.Transaction(func() error {
			if err = p.usersQ.Upsert(data.User{
				Username:  permission.Username,
				GithubId:  permission.GithubId,
				CreatedAt: time.Now(),
				AvatarUrl: permission.AvatarUrl,
			}); err != nil {
				p.log.WithError(err).Errorf("failed to create user in user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to create user in user db")
			}

			usrDb, err := p.usersQ.FilterByUsernames(permission.Username).Get()
			if err != nil {
				p.log.WithError(err).Errorf("failed to get user form user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get user from user db")
			}
			if usrDb == nil {
				p.log.Errorf("no user with such username `%s` for message action with id `%s`", permission.Username, msg.RequestId)
				return errors.Wrap(err, "no user with such username")
			}

			usersToUnverified = append(usersToUnverified, *usrDb)

			permission.UserId = usrDb.Id
			permission.Link = msg.Link
			permission.Type = msg.Type
			permission.RequestId = msg.RequestId

			err = p.permissionsQ.Upsert(permission)
			if err != nil {
				p.log.WithError(err).Errorf("failed to upsert permission for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to upsert permission in permission db")
			}

			err = p.indexHasParentChild(permission.GithubId, permission.Link)
			if err != nil {
				p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to check parent level")
			}

			return nil
		})
		if err != nil {
			p.log.WithError(err).Errorf("failed to make get users transaction for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to make get users transaction")
		}
	}

	err = p.sendUsers(msg.RequestId, usersToUnverified)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish users")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) checkHasParent(permission data.Sub) error {
	if permission.ParentId == nil {
		hasParent := false
		err := p.permissionsQ.FilterByGithubIds(permission.GithubId).
			FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			p.log.Errorf("failed to update parent level")
			return errors.Wrap(err, "failed to update parent level")
		}

		return nil
	}

	parentPermission, err := p.subsQ.WithPermissions().FilterByGithubIds(permission.GithubId).FilterByIds(*permission.ParentId).OrderBy("subs_link").Get()
	if err != nil {
		p.log.Errorf("failed to get parent permission")
		return errors.Wrap(err, "failed to get parent permission")
	}

	if parentPermission == nil || parentPermission.AccessLevel == "" {
		//suppose that it means: that user is not in parent repo only in lower level
		err = p.createHigherLevelPermissions(permission)
		if err != nil {
			return errors.Wrap(err, "failed to create higher parent permissions")
		}

		hasParent := false
		err = p.permissionsQ.FilterByGithubIds(permission.GithubId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			return errors.Wrap(err, "failed to update has parent")
		}

		return nil
	}

	err = p.permissionsQ.FilterByGithubIds(permission.GithubId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{
		ParentLink: &parentPermission.Link,
	})
	if err != nil {
		p.log.Errorf("failed to update parent link")
		return errors.Wrap(err, "failed to update parent link")
	}

	if permission.AccessLevel == parentPermission.AccessLevel {
		return nil
	}

	hasParent := false
	err = p.permissionsQ.FilterByGithubIds(permission.GithubId).
		FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
	if err != nil {
		p.log.Errorf("failed to update parent level")
		return errors.Wrap(err, "failed to update parent level")
	}

	hasChild := true
	err = p.permissionsQ.FilterByGithubIds(parentPermission.GithubId).
		FilterByLinks(parentPermission.Link).Update(data.PermissionToUpdate{HasChild: &hasChild})
	if err != nil {
		p.log.Errorf("failed to update parent level")
		return errors.Wrap(err, "failed to update parent level")
	}

	return nil
}
