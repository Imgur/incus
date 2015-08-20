package incus

import "container/list"
import "sync"
import "log"
import "time"

// A queue of redis commands
type RedisQueue struct {
	incoming chan RedisCommand
	outgoing chan RedisCommand
	pending  *list.List
	stats    RuntimeStats

	// Used to control access to the pending list
	pendingListLock sync.Mutex
	// Used to signal to exactly one goroutine that it should wake up
	pendingCond *sync.Cond
}

func NewRedisQueue(consumers int, stats RuntimeStats, pool *redisPool) *RedisQueue {
	incoming := make(chan RedisCommand)
	outgoing := make(chan RedisCommand)

	for i := 0; i < consumers; i++ {
		outgoingRcv := (<-chan RedisCommand)(outgoing)
		NewRedisQueueConsumer(outgoingRcv, pool)
	}

	pendingList := &RedisQueue{
		incoming:        incoming,
		outgoing:        outgoing,
		pending:         list.New(),
		stats:           stats,
		pendingListLock: sync.Mutex{},
		pendingCond:     sync.NewCond(&sync.Mutex{}),
	}

	go pendingList.ReceiveForever()
	go pendingList.DispatchForever()

	return pendingList
}

func (r *RedisQueue) RunAsyncTimeout(timeout time.Duration, callback RedisCallback) RedisCommandResult {
	resultChan := make(chan RedisCommandResult, 1)

	job := RedisCommand{
		Callback: callback,
		Result:   resultChan,
	}

	r.incoming <- job

	select {
	case result := <-resultChan:
		return result
	case <-time.After(timeout):
		return RedisCommandResult{
			Value: nil,
			Error: timedOut,
		}
	}
}

// Receive a command, save/enqueue it, and wake the dispatching goroutine
func (r *RedisQueue) ReceiveForever() {
	for {
		command := <-r.incoming

		r.pendingListLock.Lock()
		r.pending.PushBack(command)
		r.pendingListLock.Unlock()

		r.pendingCond.Signal()
	}
}

// Dispatch a command to a goroutine that is blocking on receiving a command.
func (r *RedisQueue) DispatchForever() {
	for {
		r.pendingCond.L.Lock()

		var pendingLength int
		var front *list.Element = nil

		for {
			r.pendingListLock.Lock()
			pendingLength = r.pending.Len()

			if DEBUG {
				log.Printf("There are %d pending commands", pendingLength)
			}

			r.stats.LogPendingRedisActivityCommandsListLength(pendingLength)

			if pendingLength != 0 {
				front = r.pending.Front()
				r.pending.Remove(front)
				r.pendingListLock.Unlock()
				break
			} else {
				r.pendingListLock.Unlock()
				r.pendingCond.Wait()
			}
		}

		r.pendingCond.L.Unlock()

		go func(value RedisCommand) {
			if DEBUG {
				log.Printf("Pushed one command to outgoing")
			}

			r.outgoing <- value
		}(front.Value.(RedisCommand))
	}
}
