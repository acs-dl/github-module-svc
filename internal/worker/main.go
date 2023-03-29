package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/data/postgres"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/processor"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

const serviceName = data.ModuleName + "-worker"

type Worker interface {
	Run(ctx context.Context)
}

type worker struct {
	logger       *logan.Entry
	processor    processor.Processor
	githubClient github.GithubClient
	linksQ       data.Links
	usersQ       data.Users
	subsQ        data.Subs
	permissionsQ data.Permissions
}

func NewWorker(cfg config.Config) Worker {
	return &worker{
		logger:       cfg.Log().WithField("runner", serviceName),
		processor:    processor.NewProcessor(cfg),
		githubClient: github.NewGithub(cfg.Github().Token, cfg.Log().WithField("runner", serviceName)),
		linksQ:       postgres.NewLinksQ(cfg.DB()),
		subsQ:        postgres.NewSubsQ(cfg.DB()),
		usersQ:       postgres.NewUsersQ(cfg.DB()),
		permissionsQ: postgres.NewPermissionsQ(cfg.DB()),
	}
}

func (w *worker) Run(ctx context.Context) {
	running.WithBackOff(
		ctx,
		w.logger,
		serviceName,
		w.processPermissions,
		15*time.Minute,
		15*time.Minute,
		15*time.Minute,
	)
}

func (w *worker) processPermissions(_ context.Context) error {
	w.logger.Info("fetching links")

	startTime := time.Now()

	links, err := w.linksQ.Select()
	if err != nil {
		return errors.Wrap(err, "failed to get links")
	}

	reqAmount := len(links)
	if reqAmount == 0 {
		w.logger.Info("no links were found")
		return nil
	}

	w.logger.Infof("found %v links", reqAmount)

	for _, link := range links {
		w.logger.Infof("processing link `%s`", link.Link)

		err = w.createSubs(link)
		if err != nil {
			w.logger.Infof("failed to create subs for link `%s", link.Link)
			return errors.Wrap(err, "failed to create subs")
		}

		w.logger.WithField("link", link.Link).Info("link was processed successfully")

	}

	err = w.removeOldUsers(startTime)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to remove old users")
		return errors.Wrap(err, "failed to remove old users")
	}

	err = w.removeOldPermissions(startTime)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to remove old permissions")
		return errors.Wrap(err, "failed to remove old permissions")
	}

	return nil
}

func (w *worker) removeOldUsers(borderTime time.Time) error {
	w.logger.Infof("started removing old users")

	users, err := w.usersQ.FilterByLowerTime(borderTime).Select()
	if err != nil {
		w.logger.Infof("failed to select users")
		return errors.Wrap(err, " failed to select users")
	}

	w.logger.Infof("found `%d` users to delete", len(users))

	for _, user := range users {
		if user.Id == nil { //if unverified user we need to remove them from `unverified-svc`
			err = w.processor.SendDeleteUser(uuid.New().String(), user)
			if err != nil {
				w.logger.WithError(err).Errorf("failed to publish delete user")
				return errors.Wrap(err, " failed to publish delete user")
			}
		}

		err = w.usersQ.Delete(user.GithubId)
		if err != nil {
			w.logger.Infof("failed to delete user with github id `%d`", user.GithubId)
			return errors.Wrap(err, " failed to delete user")
		}
	}

	w.logger.Infof("finished removing old users")
	return nil
}

func (w *worker) removeOldPermissions(borderTime time.Time) error {
	w.logger.Infof("started removing old permissions")

	permissions, err := w.permissionsQ.FilterByLowerTime(borderTime).Select()
	if err != nil {
		w.logger.Infof("failed to select permissions")
		return errors.Wrap(err, " failed to select permissions")
	}

	w.logger.Infof("found `%d` permissions to delete", len(permissions))

	for _, permission := range permissions {
		err = w.permissionsQ.Delete(permission.GithubId, permission.Type, permission.Link)
		if err != nil {
			w.logger.Infof("failed to delete permission")
			return errors.Wrap(err, " failed to delete permission")
		}
	}

	w.logger.Infof("finished removing old permissions")
	return nil
}

func (w *worker) createPermission(link string) error {
	w.logger.Infof("processing sub `%s`", link)

	if err := w.processor.HandleNewMessage(data.ModulePayload{
		RequestId: "from-worker",
		Action:    processor.GetUsersAction,
		Link:      link,
	}); err != nil {
		w.logger.Infof("failed to get users sub `%s`", link)
		return errors.Wrap(err, "failed to get users")
	}

	w.logger.Infof("successfully processed sub `%s", link)

	return nil
}

func (w *worker) createSubs(link data.Link) error {
	w.logger.Infof("creating subs for link `%s", link.Link)

	typeTo, sub, err := w.githubClient.FindType(link.Link)
	if err != nil {
		w.logger.Infof("failed to get type for link `%s`", link.Link)
		return errors.Wrap(err, "failed to get type")
	}
	if sub == nil {
		w.logger.Infof("failed to get sub for link `%s`", link.Link)
		return errors.Wrap(err, "failed to get sub")
	}

	err = w.subsQ.Upsert(data.Sub{
		Id:       sub.Id,
		Path:     sub.Path,
		Link:     sub.Link,
		Type:     typeTo,
		ParentId: nil,
	})
	if err != nil {
		w.logger.Infof("failed to upsert sub for link `%s`", link.Link)
		return errors.Wrap(err, "failed to upsert sub")
	}

	err = w.createPermission(sub.Link)
	if err != nil {
		w.logger.Infof("failed to create permissions for sub with link `%s`", link.Link)
		return errors.Wrap(err, "failed to create permissions for sub")
	}

	if typeTo == data.Repository {
		return nil
	}

	err = w.processNested(link.Link, sub.Id)
	if err != nil {
		w.logger.Infof("failed to index subs for link `%s`", link.Link)
		return errors.Wrap(err, "failed to index subs")
	}

	w.logger.Infof("finished creating subs for link `%s", link.Link)
	return nil
}

func (w *worker) processNested(link string, parentId int64) error {
	w.logger.Debugf("processing link `%s`", link)

	projects, err := w.githubClient.GetProjectsFromApi(link)
	if err != nil {
		w.logger.Infof("failed to get projects for link `%s`", link)
		return errors.Wrap(err, fmt.Sprintf("failed to get projects for link `%s`", link))
	}

	for _, project := range projects {
		err = w.subsQ.Upsert(data.Sub{
			Id:       project.Id,
			Path:     project.Path,
			Link:     link + "/" + project.Path,
			Type:     data.Repository,
			ParentId: &parentId,
		})
		if err != nil {
			w.logger.Infof("failed to upsert sub with link `%s`", link+"/"+project.Path)
			return errors.Wrap(err, fmt.Sprintf("failed to get upsert sub with link `%s`", link+"/"+project.Path))
		}

		err = w.createPermission(link + "/" + project.Path)
		if err != nil {
			w.logger.Infof("failed to create permissions for sub with link `%s`", link+"/"+project.Path)
			return errors.Wrap(err, "failed to create permissions for sub")
		}
	}

	return nil
}
