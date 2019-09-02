package fails

import (
	"log"
	"sync"

	"github.com/morikuni/failure"
)

const (
	// ErrCritical はクリティカルなエラーで少しでも大幅減点・失格になるエラー
	ErrCritical failure.StringCode = "error critical"
	// ErrApplication はアプリケーションの挙動でおかしいやつ。Verify時は1つでも失格。Validation時は一定数以上で失格
	ErrApplication failure.StringCode = "error application"
	// ErrTimeout はタイムアウトエラー。基本は大目に見る。
	ErrTimeout failure.StringCode = "error timeout"
)

type Critical struct {
	Msgs []string

	critical    int
	application int
	timeout     int

	mu sync.Mutex
}

func NewCritical() *Critical {
	msgs := make([]string, 0, 100)
	return &Critical{
		Msgs: msgs,
	}
}

func (c *Critical) GetMsgs() (msgs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Msgs[:]
}

func (c *Critical) Get() (msgs []string, critical, application, timeout int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Msgs[:], c.critical, c.application, c.timeout
}

func (c *Critical) Add(err error) {
	if err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("%+v", err)

	if msg, ok := failure.MessageOf(err); ok {
		switch code, _ := failure.CodeOf(err); code {
		case ErrCritical:
			msg += " (critical error)"
			c.critical++
		case ErrApplication:
			c.application++
		case ErrTimeout:
			msg += "（タイムアウトしました）"
			c.timeout++
		}

		c.Msgs = append(c.Msgs, msg)
	} else {
		// 想定外のエラーなのでcritical扱いにしておく
		c.critical++
		c.Msgs = append(c.Msgs, "運営に連絡してください")
	}
}
