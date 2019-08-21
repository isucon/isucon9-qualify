package main

import (
	"context"
	"encoding/json"
	"flag"
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

type Config struct {
	// PaymentPort    int
	// ShipmentPort   int
	TargetURLStr string
	TargetHost   string
}

func init() {
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flags := flag.NewFlagSet("isucon9q", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	conf := Config{}

	flags.StringVar(&conf.TargetURLStr, "target-url", "http://127.0.0.1:8000", "target url")
	flags.StringVar(&conf.TargetHost, "target-host", "isucon9.catatsuy.org", "target host")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	sp, ss, err := server.RunServer(5555, 7000)
	if err != nil {
		log.Fatal(err)
	}

	scenario.SetShipment(ss)
	scenario.SetPayment(sp)

	err = session.SetShareTargetURLs(
		conf.TargetURLStr,
		conf.TargetHost,
		"http://localhost:5555",
		"http://localhost:7000",
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("=== initialize ===")
	scenario.Initialize()
	log.Print("=== verify ===")

	cerr := scenario.Verify(context.Background())
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

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(20*time.Second))
	defer cancel()

	scenario.Validation(ctx, cerr)

	criticalMsgs = cerr.GetMsgs()
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

	log.Print("=== final check ===")

	scenario.FinalCheck(cerr)

	output := Output{
		Pass:  true,
		Score: 0,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}
