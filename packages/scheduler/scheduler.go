package scheduler

import (
	"runtime"

	"github.com/iotaledger/hive.go/workerpool"
)

var (
	inboxWorkerCount     = 1
	inboxWorkerQueueSize = 1000

	outboxWorkerCount     = runtime.GOMAXPROCS(0) * 4
	outboxWorkerQueueSize = 1000
)

var (
	InboxWorkerPool  *workerpool.WorkerPool
	OutboxWorkerPool *workerpool.WorkerPool
)
