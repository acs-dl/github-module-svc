package service

import (
	"context"
	"sync"

	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/receiver"
	"gitlab.com/distributed_lab/acs/github-module/internal/registrator"
	"gitlab.com/distributed_lab/acs/github-module/internal/sender"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/acs/github-module/internal/worker"

	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/types"
)

var availableServices = map[string]types.Runner{
	"api":      api.Run,
	"sender":   sender.Run,
	"receiver": receiver.Run,
	//"worker":    worker.Run,
	"registrar": registrator.Run,
}

func Run(cfg config.Config) {
	logger := cfg.Log().WithField("service", "main")
	ctx := context.Background()
	wg := new(sync.WaitGroup)

	logger.Info("Starting all available services...")

	stopProcessQueue := make(chan struct{})
	newPqueue := pqueue.NewPriorityQueue()
	go newPqueue.ProcessQueue(cfg.RateLimit().RequestsAmount, cfg.RateLimit().TimeLimit, stopProcessQueue)
	ctx = pqueue.CtxPQueue(newPqueue.(*pqueue.PriorityQueue), ctx)

	ctx = background.CtxConfig(cfg, ctx)

	//TODO: think about, how to save instances in context before running
	wrkr := worker.NewWorker(cfg, ctx)
	ctx = worker.CtxWorkerInstance(&wrkr, ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()

		wrkr.Run(ctx)
	}()
	logger.WithField("service", "worker").Info("Service started")
	//This is kostyl, remove it

	for serviceName, service := range availableServices {
		wg.Add(1)

		go func(name string, runner types.Runner) {
			defer wg.Done()

			runner(ctx, cfg)

		}(serviceName, service)

		logger.WithField("service", serviceName).Info("Service started")
	}

	wg.Wait()
}
