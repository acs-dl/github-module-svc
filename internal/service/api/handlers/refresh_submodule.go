package handlers

import (
	"math"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/models"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/api/requests"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/acs/github-module/internal/worker"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func RefreshSubmodule(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewRefreshSubmoduleRequest(r)
	if err != nil {
		background.Log(r).WithError(err).Error("failed to parse request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	subs, err := getSubIds(request.Data.Attributes.Links, background.SubsQ(r))
	if err != nil {
		background.Log(r).WithError(err).Error("failed to get subs")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	subsAmount := int64(len(subs))

	permissionsAmount, err := getPermissionsAmount(subs, background.SubsQ(r))
	if err != nil {
		background.Log(r).WithError(err).Error("failed to get permissions amount")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	parentContext := background.ParentContext(r.Context())

	pqueueRequestsAmount := int64(pqueue.PQueuesInstance(parentContext).SuperPQueue.Len() + pqueue.PQueuesInstance(parentContext).UsualPQueue.Len())

	requestsTimeLimit := background.Config(parentContext).RateLimit().TimeLimit
	requestsAmountLimit := background.Config(parentContext).RateLimit().RequestsAmount

	timeToHandleOneRequest := requestsTimeLimit / time.Duration(requestsAmountLimit)
	totalRequestsAmount := math.Round(float64(subsAmount+permissionsAmount+pqueueRequestsAmount) * 1.4)

	estimatedTime := time.Duration(totalRequestsAmount) * timeToHandleOneRequest

	workerInstance := worker.NewWorker(background.Config(parentContext), parentContext)
	for _, link := range request.Data.Attributes.Links {
		go func(linkToHandle string) {
			err = workerInstance.CreateSubs(linkToHandle)
			if err != nil {
				panic(err)
			}
		}(link)
	}

	w.WriteHeader(http.StatusAccepted)
	ape.Render(w, models.NewEstimatedTimeResponse(estimatedTime.String()))
}

func getPermissionsAmount(subs []data.Sub, subsQ data.Subs) (int64, error) {
	var amount = int64(0)

	for _, sub := range subs {
		permissionsAmount, err := subsQ.CountWithPermissions().FilterByIds(sub.Id).GetTotalCount()
		if err != nil {
			return -1, err
		}

		amount += permissionsAmount
	}

	return -1, nil
}

func getSubIds(links []string, subsQ data.Subs) ([]data.Sub, error) {
	subs := make([]data.Sub, 0)

	for _, link := range links {
		sub, err := subsQ.FilterByLinks(link).Get()
		if err != nil {
			return nil, err
		}

		if sub == nil {
			return nil, errors.Errorf("sub `%s` is empty", link)
		}

		subs = append(subs, *sub)

		children, err := getAllChildren([]data.Sub{*sub}, subsQ)
		if err != nil {
			return nil, err
		}

		subs = append(subs, children...)
	}

	subs = removeDuplicateSub(subs)

	return subs, nil
}

func getAllChildren(subs []data.Sub, subsQ data.Subs) ([]data.Sub, error) {
	children := make([]data.Sub, 0)

	for _, sub := range subs {
		subChildren, err := subsQ.FilterByParentIds(sub.Id).Select()
		if err != nil {
			return nil, err
		}

		if len(subChildren) == 0 {
			continue
		}
		children = append(children, subChildren...)

		nested, err := getAllChildren(subChildren, subsQ)
		if err != nil {
			return nil, err
		}

		children = append(children, nested...)

	}

	return children, nil
}

func removeDuplicateSub(arr []data.Sub) []data.Sub {
	allKeys := make(map[int64]bool)
	var list []data.Sub

	for i := range arr {
		if _, value := allKeys[arr[i].Id]; !value {
			allKeys[arr[i].Id] = true
			list = append(list, arr[i])
		}
	}

	return list
}
