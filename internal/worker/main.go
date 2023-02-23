package worker

import (
	"context"
	"fmt"
	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/data/postgres"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/processor"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"time"
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
	subsQ        data.Subs
	usersQ       data.Users
	permissionsQ data.Permissions
}

func NewWorker(cfg config.Config) Worker {
	return &worker{
		logger:       cfg.Log().WithField("runner", serviceName),
		processor:    processor.NewProcessor(cfg),
		githubClient: github.NewGithub(cfg.Github().Token),
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

	err = w.createPermissions()
	if err != nil {
		w.logger.Infof("failed to create permissions for subs")
		return errors.Wrap(err, "failed to create permissions for subs")
	}

	err = w.checkHasParent()
	if err != nil {
		w.logger.Infof("failed to check parent levels for subs")
		return errors.Wrap(err, "failed to check parent levels for for subs")
	}

	return nil
}

func (w *worker) createPermissions() error {
	w.logger.Info("fetching subs")

	subs, err := w.subsQ.Select()
	if err != nil {
		w.logger.Infof("failed to get subs")
		return errors.Wrap(err, "failed to get subs")
	}

	reqAmount := len(subs)
	if reqAmount == 0 {
		w.logger.Info("no subs were found")
		return nil
	}

	w.logger.Infof("found %v subs", reqAmount)

	for _, sub := range subs {
		w.logger.Infof("processing sub `%s`", sub.Link)

		if err = w.processor.HandleNewMessage(data.ModulePayload{
			RequestId: "from-worker",
			Action:    processor.GetUsersAction,
			Link:      sub.Link,
		}); err != nil {
			w.logger.Infof("failed to get users sub `%s`", sub.Link)
			return errors.Wrap(err, "failed to get users")
		}

		w.logger.Infof("successfully processed sub `%s", sub.Link)
	}

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
		Lpath:    fmt.Sprintf("%d", sub.Id),
	})
	if err != nil {
		w.logger.Infof("failed to upsert sub for link `%s`", link.Link)
		return errors.Wrap(err, "failed to upsert sub")
	}

	if typeTo == data.Repository {
		return nil
	}

	err = w.processNested(link.Link, sub.Id, fmt.Sprintf("%d", sub.Id))
	if err != nil {
		w.logger.Infof("failed to index subs for link `%s`", link.Link)
		return errors.Wrap(err, "failed to index subs")
	}

	w.logger.Infof("finished creating subs for link `%s", link.Link)
	return nil
}

func (w *worker) processNested(link string, parentId int64, parentLpath string) error {
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
			Lpath:    fmt.Sprintf("%s.%d", parentLpath, project.Id),
		})
		if err != nil {
			w.logger.Infof("failed to upsert sub with link `%s`", link+"/"+project.Path)
			return errors.Wrap(err, fmt.Sprintf("failed to get upsert sub with link `%s`", link+"/"+project.Path))
		}
	}

	return nil
}

func (w *worker) checkHasParent() error {
	w.logger.Infof("started checking parent levels for subs")

	users, err := w.usersQ.Select()
	if err != nil {
		w.logger.Errorf("failed to select users")
		return errors.Wrap(err, "failed to select users")
	}
	if len(users) == 0 {
		w.logger.Infof("no user was found")
		return nil
	}

	w.logger.Infof("found `%v` users to check levels", len(users))
	for _, user := range users {
		permissions, err := w.subsQ.WithPermissions().FilterByGithubIds(user.GithubId).OrderBy("subs_link").Select()
		if err != nil {
			w.logger.Errorf("failed to select permissions")
			return errors.Wrap(err, "failed to select permissions")
		}
		if len(permissions) == 0 {
			w.logger.Infof("no permission for user `%s` was found", user.Username)
			continue
		}

		w.logger.Infof("found `%v` permissions for user `%s`", len(permissions), user.Username)
		for _, permission := range permissions {
			if permission.ParentId == nil {
				err = w.permissionsQ.UpdateHasParent(data.Permission{
					HasParent: false,
					GithubId:  permission.GithubId,
					Link:      permission.Link,
				})
				if err != nil {
					w.logger.Errorf("failed to update parent level")
					return errors.Wrap(err, "failed to update parent level")
				}
				continue
			}

			parentPermission, err := w.subsQ.WithPermissions().FilterByGithubIds(user.GithubId).FilterByIds(*permission.ParentId).OrderBy("subs_link").Get()
			if err != nil {
				w.logger.Errorf("failed to get parent permission")
				return errors.Wrap(err, "failed to get parent permission")
			}

			if parentPermission == nil {
				//suppose that it means: that user is not in parent repo only in lower level
				err = w.permissionsQ.UpdateHasParent(data.Permission{
					HasParent: false,
					GithubId:  permission.GithubId,
					Link:      permission.Link,
				})
				if err != nil {
					w.logger.Errorf("failed to update parent level")
					return errors.Wrap(err, "failed to update parent level")
				}
				continue
			}

			if permission.AccessLevel != parentPermission.AccessLevel {
				err = w.permissionsQ.UpdateHasParent(data.Permission{
					HasParent: false,
					GithubId:  permission.GithubId,
					Link:      permission.Link,
				})
				if err != nil {
					w.logger.Errorf("failed to update parent level")
					return errors.Wrap(err, "failed to update parent level")
				}

				err = w.permissionsQ.UpdateHasChild(data.Permission{
					HasChild: true,
					GithubId: parentPermission.GithubId,
					Link:     parentPermission.Link,
				})
				if err != nil {
					w.logger.Errorf("failed to update parent level")
					return errors.Wrap(err, "failed to update parent level")
				}
			}
		}
		w.logger.Infof("finished checking levels for user `%s`", user.Username)
	}

	w.logger.Infof("finished checking parent levels for subs")
	return nil
}
