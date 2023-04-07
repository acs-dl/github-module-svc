package helpers

import (
	"container/heap"

	"github.com/google/uuid"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func AddFunctionInPqueue(pq *pqueue.PriorityQueue, function any, functionArgs []any, priority int) (*pqueue.QueueItem, error) {
	newUuid := uuid.New()
	queueItem := &pqueue.QueueItem{
		Uuid:     newUuid,
		Func:     function,
		Args:     functionArgs,
		Priority: priority,
	}
	heap.Push(pq, queueItem)
	item, err := pq.WaitUntilInvoked(newUuid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait until invoked")
	}

	err = pq.RemoveByUUID(newUuid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to remove by uuid")
	}

	return item, nil
}
