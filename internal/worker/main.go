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
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/processor"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

const serviceName = data.ModuleName + "-worker"

type Worker interface {
	Run(ctx context.Context)
	ProcessPermissions(ctx context.Context) error
	CreateSubs(link string) error
	GetEstimatedTime() time.Duration
}

type worker struct {
	logger        *logan.Entry
	processor     processor.Processor
	githubClient  github.GithubClient
	linksQ        data.Links
	usersQ        data.Users
	subsQ         data.Subs
	permissionsQ  data.Permissions
	pqueues       *pqueue.PQueues
	runnerDelay   time.Duration
	estimatedTime time.Duration
}

func NewWorker(cfg config.Config, ctx context.Context) Worker {
	return &worker{
		logger:        cfg.Log().WithField("runner", serviceName),
		processor:     processor.NewProcessor(cfg, ctx),
		githubClient:  github.NewGithub(cfg),
		linksQ:        postgres.NewLinksQ(cfg.DB()),
		subsQ:         postgres.NewSubsQ(cfg.DB()),
		usersQ:        postgres.NewUsersQ(cfg.DB()),
		permissionsQ:  postgres.NewPermissionsQ(cfg.DB()),
		pqueues:       pqueue.PQueuesInstance(ctx),
		estimatedTime: time.Duration(0),
		runnerDelay:   cfg.Runners().Worker,
	}
}

func (w *worker) Run(ctx context.Context) {
	running.WithBackOff(
		ctx,
		w.logger,
		serviceName,
		w.ProcessPermissions,
		w.runnerDelay,
		w.runnerDelay,
		w.runnerDelay,
	)
}

func (w *worker) ProcessPermissions(_ context.Context) error {
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

		err = w.CreateSubs(link.Link)
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

	w.estimatedTime = time.Now().Sub(startTime)
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

		err = w.usersQ.FilterByGithubIds(user.GithubId).Delete()
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
		err = w.permissionsQ.FilterByGithubIds(permission.GithubId).FilterByLinks(permission.Link).FilterByTypes(permission.Type).Delete()
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

func (w *worker) CreateSubs(link string) error {
	w.logger.Infof("creating subs for link `%s", link)

	item, err := helpers.AddFunctionInPQueue(w.pqueues.SuperPQueue, any(w.githubClient.FindType), []any{any(link)}, pqueue.LowPriority)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to add function in pqueue")
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		w.logger.WithError(err).Errorf("failed to get type")
		return errors.Wrap(err, "failed to get type")
	}
	typeSub, ok := item.Response.Value.(*github.TypeSub)
	if !ok {
		return errors.Errorf("wrong response type")
	}

	if typeSub == nil {
		w.logger.Infof("failed to get sub for link `%s`", link)
		return errors.Wrap(err, "failed to get sub")
	}

	err = w.subsQ.Upsert(data.Sub{
		Id:       typeSub.Sub.Id,
		Path:     typeSub.Sub.Path,
		Link:     typeSub.Sub.Link,
		Type:     typeSub.Type,
		ParentId: nil,
	})
	if err != nil {
		w.logger.Infof("failed to upsert sub for link `%s`", link)
		return errors.Wrap(err, "failed to upsert sub")
	}

	err = w.createPermission(typeSub.Sub.Link)
	if err != nil {
		w.logger.Infof("failed to create permissions for sub with link `%s`", link)
		return errors.Wrap(err, "failed to create permissions for sub")
	}

	if typeSub.Type == data.Repository {
		return nil
	}

	err = w.processNested(link, typeSub.Sub.Id)
	if err != nil {
		w.logger.Infof("failed to index subs for link `%s`", link)
		return errors.Wrap(err, "failed to index subs")
	}

	w.logger.Infof("finished creating subs for link `%s", link)
	return nil
}

func (w *worker) processNested(link string, parentId int64) error {
	w.logger.Debugf("processing link `%s`", link)

	item, err := helpers.AddFunctionInPQueue(w.pqueues.SuperPQueue, any(w.githubClient.GetProjectsFromApi), []any{any(link)}, pqueue.LowPriority)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to add function in pqueue")
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		w.logger.Infof("failed to get projects for link `%s`", link)
		return errors.Wrap(err, fmt.Sprintf("failed to get projects for link `%s`", link))
	}

	projects, ok := item.Response.Value.([]data.Sub)
	if !ok {
		return errors.Errorf("wrong response type")
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

func (w *worker) GetEstimatedTime() time.Duration {
	return w.estimatedTime
}

func WorkerInstance(ctx context.Context) *Worker {
	return ctx.Value(background.WorkerCtxKey).(*Worker)
}

func CtxWorkerInstance(entry *Worker, ctx context.Context) context.Context {
	return context.WithValue(ctx, background.WorkerCtxKey, entry)
}
