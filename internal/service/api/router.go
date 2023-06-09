package api

import (
	"fmt"

	auth "github.com/acs-dl/auth-svc/middlewares"
	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/data/postgres"
	"github.com/acs-dl/github-module-svc/internal/service/api/handlers"
	"github.com/acs-dl/github-module-svc/internal/service/background"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (r *Router) apiRouter() chi.Router {
	router := chi.NewRouter()

	logger := r.cfg.Log().WithField("service", fmt.Sprintf("%s-api", data.ModuleName))

	secret := r.cfg.JwtParams().Secret

	router.Use(
		ape.RecoverMiddleware(logger),
		ape.LoganMiddleware(logger),
		ape.CtxMiddleware(
			//base
			background.CtxLog(logger),

			// storage
			background.CtxPermissionsQ(postgres.NewPermissionsQ(r.cfg.DB())),
			background.CtxUsersQ(postgres.NewUsersQ(r.cfg.DB())),
			background.CtxLinksQ(postgres.NewLinksQ(r.cfg.DB())),
			background.CtxSubsQ(postgres.NewSubsQ(r.cfg.DB())),

			// other configs
			background.CtxParentContext(r.parentContext),
		),
	)

	router.Route("/integrations/github", func(r chi.Router) {
		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Get("/get_input", handlers.GetInputs)
		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Get("/get_available_roles", handlers.GetRoles)

		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Route("/links", func(r chi.Router) {
				r.Post("/", handlers.AddLink)
				r.Delete("/", handlers.RemoveLink)
			})

		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Get("/permissions", handlers.GetPermissions)

		r.Get("/role", handlers.GetRole)               // comes from orchestrator
		r.Get("/roles", handlers.GetRolesMap)          // comes from orchestrator
		r.Get("/user_roles", handlers.GetUserRolesMap) // comes from orchestrator

		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Route("/estimate_refresh", func(r chi.Router) {
				r.Post("/submodule", handlers.GetEstimatedRefreshSubmodule)
				r.Post("/module", handlers.GetEstimatedRefreshModule)
			})

		r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
			Get("/submodule", handlers.CheckSubmodule)

		r.Route("/users", func(r chi.Router) {
			r.Get("/{id}", handlers.GetUserById) // comes from orchestrator

			r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
				Get("/", handlers.GetUsers)
			r.With(auth.Jwt(secret, data.ModuleName, []string{data.Roles["read"], data.Roles["triage"], data.Roles["write"], data.Roles["maintain"], data.Roles["admin"], data.Roles["member"]}...)).
				Get("/unverified", handlers.GetUnverifiedUsers)
		})
	})

	return router
}
