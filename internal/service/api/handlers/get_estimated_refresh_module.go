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

func GetEstimatedRefreshModule(w http.ResponseWriter, r *http.Request) {
	parentContext := background.ParentContext(r.Context())
	workerInstance := *worker.WorkerInstance(parentContext)

	pqueueRequestsAmount := int64(pqueue.PQueuesInstance(parentContext).SuperUserPQueue.Len() + pqueue.PQueuesInstance(parentContext).UserPQueue.Len())
	requestsTimeLimit := background.Config(parentContext).RateLimit().TimeLimit
	requestsAmountLimit := background.Config(parentContext).RateLimit().RequestsAmount

	timeToHandleOneRequest := requestsTimeLimit / time.Duration(requestsAmountLimit)
	estimatedTime := time.Duration(pqueueRequestsAmount)*timeToHandleOneRequest + workerInstance.GetEstimatedTime()

	ape.Render(w, models.NewEstimatedTimeResponse(estimatedTime.String()))
}