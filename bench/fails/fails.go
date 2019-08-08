package fails

import (
	"fmt"
	"os"
	"sync"
)

type Error struct {
	Msg string
	Err error
}

func (e *Error) Error() string {
	return e.Msg + ": " + e.Err.Error()
}

func NewError(err error, msg string) *Error {
	ferr := &Error{
		Msg: msg,
		Err: err,
	}
	fmt.Fprintln(os.Stderr, ferr.Error())

	return ferr
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

	if ferr, ok := err.(*Error); ok {
		c.Msgs = append(c.Msgs, ferr.Msg)
	} else {
		c.Msgs = append(c.Msgs, "運営に連絡してください")
	}
}
