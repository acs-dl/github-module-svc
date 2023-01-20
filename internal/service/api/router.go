package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/data/postgres"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/handlers"
	"gitlab.com/distributed_lab/ape"
)

func (r *apiRouter) apiRouter() chi.Router {
	router := chi.NewRouter()

	logger := r.cfg.Log().WithField("service", fmt.Sprintf("%s-api", data.ModuleName))

	router.Use(
		ape.RecoverMiddleware(logger),
		ape.LoganMiddleware(logger),
		ape.CtxMiddleware(
			//base
			handlers.CtxLog(logger),

			// storage
			handlers.CtxPermissionsQ(postgres.NewPermissionsQ(r.cfg.DB())),
			handlers.CtxUsersQ(postgres.NewUsersQ(r.cfg.DB())),
			handlers.CtxLinksQ(postgres.NewLinksQ(r.cfg.DB())),

			// connectors

			// other configs
			handlers.CtxParams(r.cfg.Github()),
		),
	)

	router.Route("/integrations/gitlab", func(r chi.Router) {
		r.Get("/get_input", handlers.GetInputs)
		r.Get("/get_available_roles", handlers.GetRoles)

		r.Route("/links", func(r chi.Router) {
			r.Post("/", handlers.AddLink)
			r.Delete("/", handlers.RemoveLink)
		})

		r.Get("/permissions", handlers.GetPermissions)

		r.Route("/users/{id}", func(r chi.Router) {
			r.Get("/permissions", handlers.GetUserPermissions)
		})
	})

	return router
}
