package processor

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateGetUsers(msg data.ModulePayload) error {
	return validation.Errors{
		"link": validation.Validate(msg.Link, validation.Required),
	}.Filter()
}

func (p *processor) handleGetUsersAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateGetUsers(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	item, err := helpers.AddFunctionInPqueue(p.pqueue, any(p.githubClient.FindType), []any{any(msg.Link)}, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to get type for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get type")
	}
	typeSub, ok := item.Response.Value.(*github.TypeSub)
	if !ok {
		return errors.Errorf("wrong response type")
	}

	if typeSub == nil {
		p.log.WithError(err).Errorf("type not found for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "type not found")
	}
	msg.Type = typeSub.Type

	if validation.Validate(msg.Type, validation.Required, validation.In(data.Organization, data.Repository)) != nil {
		p.log.WithError(err).Errorf("unexpected link type `%s` for message action with id `%s`", msg.Type, msg.RequestId)
		return errors.Wrap(err, "something wrong with link type")
	}

	item, err = helpers.AddFunctionInPqueue(p.pqueue, any(p.githubClient.GetUsersFromApi), []any{any(msg.Link), any(msg.Type)}, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to get users from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting users from api")
	}
	users, ok := item.Response.Value.([]data.Permission)
	if !ok {
		return errors.Errorf("wrong response type")
	}

	usersToUnverified := make([]data.User, 0)

	for _, user := range users {
		//api doesn't return role for organization members
		if msg.Type == data.Organization {
			item, err = helpers.AddFunctionInPqueue(p.pqueue, any(p.githubClient.CheckOrgCollaborator), []any{any(msg.Link), any(user.Username)}, pqueue.LowPriority)
			if err != nil {
				p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to add function in pqueue")
			}

			err = item.Response.Error
			if err != nil {
				p.log.WithError(err).Errorf("failed to get permission from api for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get permission from api")
			}
			checkPermission, ok := item.Response.Value.(*github.CheckPermission)
			if !ok {
				return errors.Errorf("wrong response type")
			}
			if checkPermission == nil {
				p.log.Errorf("something went wrong with getting permission for message action with id `%s`", msg.RequestId)
				return errors.Errorf("something went wrong with getting permission from api")
			}

			user.AccessLevel = checkPermission.Permission.AccessLevel
		}

		err = p.managerQ.Transaction(func() error {
			if err = p.usersQ.Upsert(data.User{
				Username:  user.Username,
				GithubId:  user.GithubId,
				CreatedAt: time.Now(),
				AvatarUrl: user.AvatarUrl,
			}); err != nil {
				p.log.WithError(err).Errorf("failed to create user in user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to create user in user db")
			}

			p.resetFilters()

			usrDb, err := p.usersQ.FilterByUsernames(user.Username).Get()
			if err != nil {
				p.log.WithError(err).Errorf("failed to get user form user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get user from user db")
			}
			if usrDb == nil {
				p.log.Errorf("no user with such username `%s` for message action with id `%s`", user.Username, msg.RequestId)
				return errors.Wrap(err, "no user with such username")
			}

			usersToUnverified = append(usersToUnverified, *usrDb)

			user.UserId = usrDb.Id
			user.Link = msg.Link
			user.Type = msg.Type
			user.RequestId = msg.RequestId

			err = p.permissionsQ.Upsert(user)
			if err != nil {
				p.log.WithError(err).Errorf("failed to upsert permission for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to upsert permission in permission db")
			}

			subPermission, err := p.subsQ.WithPermissions().FilterByGithubIds(user.GithubId).FilterByLinks(user.Link).OrderBy("subs_link").Get()
			if err != nil {
				p.log.WithError(err).Errorf("failed to get permission for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get permission in permission db")
			}
			if subPermission == nil {
				p.log.WithError(err).Errorf("got empty permission for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "got empty permission in permission db")
			}

			err = p.checkHasParent(*subPermission)
			if err != nil {
				p.log.WithError(err).Errorf("failed to check parent level for message action with id `%s`", msg.RequestId)
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

	p.resetFilters()
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

	if parentPermission == nil {
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
