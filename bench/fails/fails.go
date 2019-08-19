package fails

import (
	"log"
	"sync"

	"github.com/morikuni/failure"
)

type Error struct {
	Msg string
	Err error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Err.Error()
}

type Critical struct {
	Msgs []string
	mu   sync.Mutex
}

func NewCritical() *Critical {
	msgs := make([]string, 0, 100)
	return &Critical{
		Msgs: msgs,
	}
}

func (c *Critical) GetMsgs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Msgs[:]
}

func (c *Critical) Add(err error) {
	if err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("%+v", err)

	if msg, ok := failure.MessageOf(err); ok {
		c.Msgs = append(c.Msgs, msg)
	} else {
		c.Msgs = append(c.Msgs, "運営に連絡してください")
	}
}
