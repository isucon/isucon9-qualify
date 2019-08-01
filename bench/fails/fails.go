package fails

import (
	"fmt"
	"os"
	"sync"
)

var mu sync.Mutex
var messages []string
var isCritical bool

func init() {
	messages = make([]string, 100)
}

func Get() []string {
	mu.Lock()
	allMessages := messages[:]
	mu.Unlock()
	return allMessages
}

func Add(msg string, err error) {
	mu.Lock()
	messages = append(messages, msg)
	mu.Unlock()

	if err != nil {
		msg += " error: " + err.Error()
	}
	fmt.Fprintln(os.Stderr, msg)
}

type Logger struct {
}

func (l *Logger) Add(msg string, err error) {
	Add(msg, err)
}
