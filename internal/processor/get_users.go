package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"time"
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

	msg.Type, _, err = p.githubClient.FindType(msg.Link)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get type for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get type")
	}

	if validation.Validate(msg.Type, validation.Required, validation.In(data.Organization, data.Repository)) != nil {
		p.log.WithError(err).Errorf("unexpected link type `%s` for message action with id `%s`", msg.Type, msg.RequestId)
		return errors.Wrap(err, "something wrong with link type")
	}

	users, err := p.githubClient.GetUsersFromApi(msg.Link, msg.Type)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get users from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting users from api")
	}

	usersToUnverified := make([]data.User, 0)

	for _, user := range users {
		//api doesn't return role for organization members
		if msg.Type == data.Organization {
			_, permission, err := p.githubClient.CheckOrgCollaborator(msg.Link, user.Username)
			if err != nil {
				p.log.WithError(err).Errorf("failed to get permission from api for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get permission from api")
			}
			if permission == nil {
				p.log.Errorf("something went wrong with getting permission for message action with id `%s`", msg.RequestId)
				return errors.Errorf("something went wrong with getting permission from api")
			}
			user.AccessLevel = permission.AccessLevel
		}

		dbUser := data.User{
			Username:  user.Username,
			GithubId:  user.GithubId,
			CreatedAt: time.Now(),
			AvatarUrl: user.AvatarUrl,
		}
		usersToUnverified = append(usersToUnverified, dbUser)

		err = p.managerQ.Transaction(func() error {
			if err = p.usersQ.Upsert(dbUser); err != nil {
				p.log.WithError(err).Errorf("failed to create user in user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to create user in user db")
			}

			p.usersQ.ResetFilters()

			usrDb, err := p.usersQ.GetByUsername(user.Username)
			if err != nil {
				p.log.WithError(err).Errorf("failed to get user form user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get user from user db")
			}
			if usrDb == nil {
				p.log.Errorf("no user with such username `%s` for message action with id `%s`", user.Username, msg.RequestId)
				return errors.Wrap(err, "no user with such username")
			}

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
		err := p.permissionsQ.UpdateHasParent(data.Permission{
			HasParent: false,
			GithubId:  permission.GithubId,
			Link:      permission.Link,
		})
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
		err = p.permissionsQ.UpdateHasParent(data.Permission{
			HasParent: false,
			GithubId:  permission.GithubId,
			Link:      permission.Link,
		})
		if err != nil {
			p.log.Errorf("failed to update parent level")
			return errors.Wrap(err, "failed to update parent level")
		}

		return nil
	}

	err = p.permissionsQ.UpdateParentLink(data.Permission{
		GithubId:   permission.GithubId,
		Link:       permission.Link,
		ParentLink: &parentPermission.Link,
	})
	if err != nil {
		p.log.Errorf("failed to update parent link")
		return errors.Wrap(err, "failed to update parent link")
	}

	if permission.AccessLevel == parentPermission.AccessLevel {
		return nil
	}

	err = p.permissionsQ.UpdateHasParent(data.Permission{
		HasParent: false,
		GithubId:  permission.GithubId,
		Link:      permission.Link,
	})
	if err != nil {
		p.log.Errorf("failed to update parent level")
		return errors.Wrap(err, "failed to update parent level")
	}

	err = p.permissionsQ.UpdateHasChild(data.Permission{
		HasChild: true,
		GithubId: parentPermission.GithubId,
		Link:     parentPermission.Link,
	})
	if err != nil {
		p.log.Errorf("failed to update parent level")
		return errors.Wrap(err, "failed to update parent level")
	}

	return nil
}
