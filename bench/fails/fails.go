package fails

import (
	"log"
	"sync"

	"github.com/morikuni/failure"
)

const (
	ErrApplication failure.StringCode = "error application"
	ErrSession     failure.StringCode = "error session"
	ErrTimeout     failure.StringCode = "error timeout"
)

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
		switch c, _ := failure.CodeOf(err); c {
		case ErrTimeout:
			msg += "（タイムアウトしました）"
		}

		c.Msgs = append(c.Msgs, msg)
	} else {
		c.Msgs = append(c.Msgs, "運営に連絡してください")
	}
}
