package fails

import (
	"log"
	"sync"
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

func NewError(err error, msg string) *Error {
	ferr := &Error{
		Msg: msg,
		Err: err,
	}
	if err != nil {
		log.Printf("%s: %+v", msg, err)
	} else {
		log.Print(ferr)
	}

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
		log.Printf("%+v", err)

		c.Msgs = append(c.Msgs, "運営に連絡してください")
	}
}
