package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/scenario"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
)

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Campaign int      `json:"campaign"`
	Language string   `json:"language"`
	Messages []string `json:"messages"`
}

type Config struct {
	TargetURLStr string
	TargetHost   string
	ShipmentURL  string
	PaymentURL   string
	PaymentPort  int
	ShipmentPort int

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
	dataDir := ""
	staticDir := ""

	flags.StringVar(&conf.TargetURLStr, "target-url", "http://127.0.0.1:8000", "target url")
	flags.StringVar(&conf.TargetHost, "target-host", "isucon9.catatsuy.org", "target host")
	flags.StringVar(&conf.PaymentURL, "payment-url", "http://localhost:5555", "payment url")
	flags.StringVar(&conf.ShipmentURL, "shipment-url", "http://localhost:7000", "shipment url")
	flags.IntVar(&conf.PaymentPort, "payment-port", 5555, "payment service port")
	flags.IntVar(&conf.ShipmentPort, "shipment-port", 7000, "shipment service port")
	flags.StringVar(&dataDir, "data-dir", "initial-data", "data directory")
	flags.StringVar(&staticDir, "static-dir", "webapp/public/static", "static file directory")
	flags.StringVar(&allowedIPStr, "allowed-ips", "", "allowed ips (comma separated)")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if allowedIPStr != "" {
		for _, str := range strings.Split(allowedIPStr, ",") {
			aip := net.ParseIP(str)
			if aip == nil {
				log.Fatalf("allowed-ips: %s cannot be parsed", str)
			}
			conf.AllowedIPs = append(conf.AllowedIPs, aip)
		}
	}

	// 外部サービスの起動
	sp, ss, err := server.RunServer(conf.PaymentPort, conf.ShipmentPort, dataDir, conf.AllowedIPs)
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

	// 初期データの準備
	asset.Initialize(dataDir, staticDir)
	scenario.InitSessionPool()

	log.Print("=== initialize ===")
	// 初期化：/initialize にリクエストを送ることで、外部リソースのURLを指定する・DBのデータを初期データのみにする
	campaign, language := scenario.Initialize(context.Background(), session.ShareTargetURLs.PaymentURL.String(), session.ShareTargetURLs.ShipmentURL.String())
	eMsgs := fails.ErrorsForCheck.GetMsgs()
	if len(eMsgs) > 0 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: eMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	log.Print("=== verify ===")
	// 初期チェック：正しく動いているかどうかを確認する
	// 明らかにおかしいレスポンスを返しているアプリケーションはさっさと停止させることで、運営側のリソースを使い果たさない・他サービスへの攻撃に利用されるを防ぐ
	scenario.Verify(context.Background())
	eMsgs = fails.ErrorsForCheck.GetMsgs()
	if len(eMsgs) > 0 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: eMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(scenario.ExecutionSeconds*time.Second))
	defer cancel()

	log.Print("=== validation ===")

	// 外部サービスのレイテンシを追加
	// verify時にもレイテンシを入れていると時間がかかるので、Validationで入れる
	ss.SetDelay(800 * time.Millisecond)
	sp.SetDelay(800 * time.Millisecond)

	// 一番大切なメイン処理：checkとloadの大きく2つの処理を行う
	// checkはアプリケーションが正しく動いているか常にチェックする
	// 理想的には全リクエストはcheckされるべきだが、それをやるとパフォーマンスが出し切れず、最適化されたアプリケーションよりも遅くなる
	// checkとloadは区別がつかないようにしないといけない。loadのリクエストはログアウト状態しかなかったので、ログアウト時のキャッシュを強くするだけでスコアがはねる問題が過去にあった
	// 今回はほぼ全リクエストがログイン前提になっているので、checkとloadの区別はできないはず
	scenario.Validation(ctx, campaign)

	// context.Canceledのエラーは直後に取れば基本的には入ってこない
	eMsgs, cCnt, aCnt, tCnt := fails.ErrorsForCheck.Get()
	// critical errorは1つでもあれば、application errorは10回以上で失格
	if cCnt > 0 || aCnt >= 10 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: uniqMsgs(eMsgs),
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	<-time.After(1 * time.Second)

	log.Print("=== final check ===")
	// 最終チェック：ベンチマーカーの記録とアプリケーションの記録を突き合わせて、最終的なスコアを算出する
	score := scenario.FinalCheck(context.Background())

	// application errorだけが発生する
	fMsgs, _, faCnt, _ := fails.ErrorsForFinal.Get()
	msgs := append(uniqMsgs(eMsgs), fMsgs...)

	aCnt += faCnt

	// application errorは10回以上で失格
	if aCnt >= 10 {
		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: msgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	// application errorは1回で500点減点
	penalty := int64(500 * aCnt)

	if tCnt > 200 {
		// trivial errorは200回を超えたら100回毎に5000点減点
		penalty += int64(5000 * (1 + (tCnt-200)/100))
	}

	log.Print(score, penalty)

	score -= penalty

	// 0点以下なら失格
	if score <= 0 {
		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: msgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	output := Output{
		Pass:     true,
		Score:    score,
		Campaign: campaign,
		Language: language,
		Messages: msgs,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}

func uniqMsgs(allMsgs []string) []string {
	sort.Strings(allMsgs)
	msgs := make([]string, 0, len(allMsgs))

	tmp := ""

	// 適当にuniqする
	for _, m := range allMsgs {
		if tmp != m {
			tmp = m
			msgs = append(msgs, m)
		}
	}

	return msgs
}
