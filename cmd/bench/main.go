package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/isucon/isucon9-qualify/bench/scenario"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
)

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Messages []string `json:"messages"`
}

func init() {
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	err := server.RunServer(5555, 7000)
	if err != nil {
		log.Fatal(err)
	}

	err = session.SetShareTargetURLs(
		"http://localhost:8000",
		"http://localhost:5555",
		"http://localhost:7000",
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("=== initialize ===")
	scenario.Initialize()
	log.Print("=== verify ===")

	cerr := scenario.Verify()
	criticalMsgs := cerr.GetMsgs()
	if len(criticalMsgs) > 0 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Messages: criticalMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	log.Print("=== validation ===")

	output := Output{
		Pass:  true,
		Score: 0,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}
