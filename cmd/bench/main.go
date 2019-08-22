package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
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
	ShipmentURL  string
	PaymentURL   string

	AllowedIPs []net.IP
}

func init() {
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flags := flag.NewFlagSet("isucon9q", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	conf := Config{}
	allowedIPStr := ""

	flags.StringVar(&conf.TargetURLStr, "target-url", "http://127.0.0.1:8000", "target url")
	flags.StringVar(&conf.TargetHost, "target-host", "isucon9.catatsuy.org", "target host")
	flags.StringVar(&conf.PaymentURL, "payment-url", "http://localhost:5555", "payment url")
	flags.StringVar(&conf.ShipmentURL, "shipment-url", "http://localhost:7000", "shipment url")
	flags.StringVar(&allowedIPStr, "allowed-ips", "", "allowed ips (comma separated)")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	for _, str := range strings.Split(allowedIPStr, ",") {
		aip := net.ParseIP(str)
		if aip == nil {
			log.Printf("%s cannot be parsed", str)
			continue
		}
		conf.AllowedIPs = append(conf.AllowedIPs, aip)
	}

	sp, ss, err := server.RunServer(5555, 7000, conf.AllowedIPs)
	if err != nil {
		log.Fatal(err)
	}

	scenario.SetShipment(ss)
	scenario.SetPayment(sp)

	err = session.SetShareTargetURLs(
		conf.TargetURLStr,
		conf.TargetHost,
		conf.PaymentURL,
		conf.ShipmentURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("=== initialize ===")
	scenario.Initialize(session.ShareTargetURLs.PaymentURL.String(), session.ShareTargetURLs.ShipmentURL.String())
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
