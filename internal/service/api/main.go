package api

import (
	"context"
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type apiRouter struct {
	cfg           config.Config
	parentContext context.Context
}

func (r *apiRouter) run() error {
	router := r.apiRouter()

	if err := r.cfg.Copus().RegisterChi(router); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	return http.Serve(r.cfg.Listener(), router)
}

func NewApiRouter(ctx context.Context, cfg config.Config) *apiRouter {
	return &apiRouter{cfg: cfg, parentContext: ctx}
}

func Run(ctx context.Context, cfg config.Config) {
	if err := NewApiRouter(ctx, cfg).run(); err != nil {
		panic(err)
	}
}
