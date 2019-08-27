package scenario

import (
	"sync"

	"github.com/isucon/isucon9-qualify/bench/session"
)

var (
	ActiveSellerPool *Queue
	BuyerPool        *Queue
)

type Queue struct {
	sync.Mutex

	items []*session.Session
}

func InitSessionPool() {
	ActiveSellerPool = NewQueue()
	BuyerPool = NewQueue()
}

func NewQueue() *Queue {
	m := make([]*session.Session, 0, 100)
	q := &Queue{
		items: m,
	}
	return q
}

func (q *Queue) Enqueue(s *session.Session) {
	q.Lock()
	defer q.Unlock()

	q.items = append(q.items, s)
}

func (q *Queue) Dequeue() *session.Session {
	q.Lock()
	defer q.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	s := q.items[0]
	q.items = q.items[1:]

	return s
}

func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()

	return len(q.items)
}
