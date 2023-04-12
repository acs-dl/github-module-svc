package handlers

import (
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/acs/github-module/internal/worker"
	"gitlab.com/distributed_lab/ape"
)

func RefreshModule(w http.ResponseWriter, r *http.Request) {
	parentContext := background.ParentContext(r.Context())

	workerInstance := *worker.WorkerInstance(parentContext)

	pqueueRequestsAmount := int64(pqueue.PQueuesInstance(parentContext).SuperPQueue.Len() + pqueue.PQueuesInstance(parentContext).UsualPQueue.Len())
	requestsTimeLimit := background.Config(parentContext).RateLimit().TimeLimit
	requestsAmountLimit := background.Config(parentContext).RateLimit().RequestsAmount

	timeToHandleOneRequest := requestsTimeLimit / time.Duration(requestsAmountLimit)
	estimatedTime := time.Duration(pqueueRequestsAmount)*timeToHandleOneRequest + workerInstance.GetEstimatedTime()

	go func() {
		err := workerInstance.ProcessPermissions(parentContext)
		if err != nil {
			panic(err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	ape.Render(w, models.NewEstimatedTimeResponse(estimatedTime.String()))
}
