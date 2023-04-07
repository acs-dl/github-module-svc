package worker

import (
	"context"

	"gitlab.com/distributed_lab/acs/github-module/internal/config"
)

func Run(ctx context.Context, cfg config.Config) {
	NewWorker(cfg, ctx).Run(ctx)
}
