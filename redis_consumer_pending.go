package incus

import "container/list"
import "sync"
import "log"

type RedisConsumerPendingList struct {
	incoming <-chan RedisCommand
	outgoing chan RedisCommand
	pending  *list.List
	lock     sync.Mutex
	stats    RuntimeStats

	// Used to signal to exactly one goroutine that it should wait up
	pendingCond *sync.Cond
}

func NewRedisConsumerPendingList(consumers int, incoming <-chan RedisCommand, stats RuntimeStats, pool *redisPool) *RedisConsumerPendingList {
	outgoing := make(chan RedisCommand)

	for i := 0; i < consumers; i++ {
		outgoingRcv := (<-chan RedisCommand)(outgoing)
		NewRedisConsumer(outgoingRcv, pool)
	}

	pendingList := &RedisConsumerPendingList{
		incoming:    incoming,
		outgoing:    outgoing,
		pending:     list.New(),
		lock:        sync.Mutex{},
		stats:       stats,
		pendingCond: sync.NewCond(&sync.Mutex{}),
	}

	go pendingList.ConsumeForever()
	go pendingList.ForwardForever()

	return pendingList
}

func (r *RedisConsumerPendingList) ConsumeForever() {
	for {
		command := <-r.incoming
		if DEBUG {
			log.Println("Dequeued command in pending")
		}
		r.lock.Lock()
		r.pending.PushBack(command)
		r.lock.Unlock()

		r.pendingCond.Signal()
	}
}

func (r *RedisConsumerPendingList) ForwardForever() {
	for {
		r.pendingCond.L.Lock()

		var pendingLength int
		var front *list.Element = nil

		for {
			r.lock.Lock()
			pendingLength = r.pending.Len()

			if DEBUG {
				log.Printf("There are %d pending commands", pendingLength)
			}

			r.stats.LogPendingRedisActivityCommandsListLength(pendingLength)

			if pendingLength != 0 {
				front = r.pending.Front()
				r.pending.Remove(front)
				r.lock.Unlock()
				break
			} else {
				r.lock.Unlock()
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
