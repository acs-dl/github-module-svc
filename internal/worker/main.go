package worker

import (
	"context"
	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/data/postgres"
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
	permissionsQ data.Permissions
	linksQ       data.Links
}

func NewWorker(cfg config.Config) Worker {
	return &worker{
		logger:       cfg.Log().WithField("runner", serviceName),
		processor:    processor.NewProcessor(cfg),
		permissionsQ: postgres.NewPermissionsQ(cfg.DB()),
		linksQ:       postgres.NewLinksQ(cfg.DB()),
	}
}

func (w *worker) Run(ctx context.Context) {
	running.WithBackOff(
		ctx,
		w.logger,
		serviceName,
		w.processPermissions,
		5*time.Minute,
		5*time.Minute,
		5*time.Minute,
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

		if err = w.processor.HandleNewMessage(data.ModulePayload{
			RequestId: "from-worker",
			Action:    processor.GetUsersAction,
			Link:      link.Link,
		}); err != nil {
			w.logger.Infof("failed to get users for link `%s`", link.Link)
			return errors.Wrap(err, "failed to get users")
		}

		w.logger.WithField("link", link.Link).Info("link was processed successfully")
	}

	return nil
}
