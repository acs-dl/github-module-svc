package processor

import (
	"context"

	"gitlab.com/distributed_lab/logan/v3"

	"github.com/acs-dl/github-module-svc/internal/config"
	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/data/manager"
	"github.com/acs-dl/github-module-svc/internal/data/postgres"
	"github.com/acs-dl/github-module-svc/internal/github"
	"github.com/acs-dl/github-module-svc/internal/pqueue"
	"github.com/acs-dl/github-module-svc/internal/sender"
)

const (
	ServiceName = data.ModuleName + "-processor"

	SetUsersAction    = "set_users"
	DeleteUsersAction = "delete_users"
)

type Processor interface {
	HandleGetUsersAction(msg data.ModulePayload) error
	HandleAddUserAction(msg data.ModulePayload) error
	HandleUpdateUserAction(msg data.ModulePayload) error
	HandleRemoveUserAction(msg data.ModulePayload) error
	HandleDeleteUserAction(msg data.ModulePayload) error
	HandleVerifyUserAction(msg data.ModulePayload) error
	SendDeleteUser(uuid string, user data.User) error
}

type processor struct {
	log             *logan.Entry
	githubClient    github.GithubClient
	permissionsQ    data.Permissions
	subsQ           data.Subs
	usersQ          data.Users
	managerQ        *manager.Manager
	sender          *sender.Sender
	pqueues         *pqueue.PQueues
	unverifiedTopic string
}

func NewProcessorAsInterface(cfg config.Config, ctx context.Context) interface{} {
	return interface{}(&processor{
		log:             cfg.Log().WithField("service", ServiceName),
		githubClient:    github.GithubClientInstance(ctx),
		sender:          sender.SenderInstance(ctx),
		pqueues:         pqueue.PQueuesInstance(ctx),
		managerQ:        manager.NewManager(cfg.DB()),
		permissionsQ:    postgres.NewPermissionsQ(cfg.DB()),
		subsQ:           postgres.NewSubsQ(cfg.DB()),
		usersQ:          postgres.NewUsersQ(cfg.DB()),
		unverifiedTopic: cfg.Amqp().Unverified,
	})
}

func ProcessorInstance(ctx context.Context) Processor {
	return ctx.Value(ServiceName).(Processor)
}

func CtxProcessorInstance(entry interface{}, ctx context.Context) context.Context {
	return context.WithValue(ctx, ServiceName, entry)
}
