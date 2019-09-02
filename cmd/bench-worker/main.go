package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Server struct {
	ID            int    `json:"id"`
	Hostname      string `json:"hostname"`
	GlobalIP      string `json:"global_ip"`
	PrivateIP     string `json:"private_ip"`
	IsBenchTarget bool   `json:"is_bench_target"`
}

type Team struct {
	ID      int       `json:"id"`
	Owner   int       `json:"owner"`
	Name    string    `json:"name"`
	Servers []*Server `json:"servers"`
}

type Job struct {
	ID     int    `json:"id"`
	Team   *Team  `json:"team"`
	Status string `json:"status"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type Result struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Score    int    `json:"score"`
	IsPassed bool   `json:"is_passed"`
	Reason   string `json:"reason"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type JobResult struct {
	Stdout string
	Stderr string
}

type JobResultStdout struct {
	Pass     bool     `json:"pass"`
	Score    int      `json:"score"`
	Messages []string `json:"messages"`
}

const (
	apiEndpointDev         = "http://portal-dev.isucon9.hinatan.net"
	defaultInterval        = 3 * time.Second
	maxStderrLength        = 8 * 1024 * 1024
	maxNumMessage          = 20
	maxBenchmarkTime       = 240 * time.Second
	defaultBenchmarkerPath = "/home/isucon/isucari/bin/benchmarker"
)

var (
	apiClient        *http.Client
	errorJobNotFound = fmt.Errorf("Job not found")
)

func init() {
	apiClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func dequeue(ep string) (*Job, error) {
	uri := fmt.Sprintf("%s/internal/job/dequeue/", ep)
	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errorJobNotFound
	}

	job := Job{}
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func joinN(messages []string, n int) string {
	if len(messages) > n {
		strings.Join(messages[:n], ",\n")
	}
	return strings.Join(messages, ",\n")
}

func report(ep string, job *Job, jobResult *JobResult) error {
	status := "done"
	var jobResultStdout JobResultStdout
	if err := json.NewDecoder(strings.NewReader(jobResult.Stdout)).Decode(&jobResultStdout); err != nil {
		jobResultStdout = JobResultStdout{
			Pass: false,
			Score: 0,
			Messages: []string{"運営に連絡してください"},
		}
		status = "aborted"
	}

	result := Result{
		ID:       job.ID,
		Status:   status,
		Score:    jobResultStdout.Score,
		IsPassed: jobResultStdout.Pass,
		Reason:   joinN(jobResultStdout.Messages, maxNumMessage),
		Stdout:   jobResult.Stdout,
		Stderr:   jobResult.Stderr,
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(result); err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/internal/job/%d/report/", ep, job.ID)
	req, err := http.NewRequest(http.MethodPost, uri, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

func findBenchmarkTargetServer(job *Job) (*Server, error) {
	for _, server := range job.Team.Servers {
		if server.IsBenchTarget {
			return server, nil
		}
	}
	return nil, fmt.Errorf("benchmark target server not found")
}

func getExternalServiceSuffix() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(hostname, "bench"), nil
}

func runBenchmarker(benchmarkerPath string, job *Job) (*JobResult, error) {
	target, err := findBenchmarkTargetServer(job)
	if err != nil {
		return &JobResult{}, err
	}
	allowedIPs := []string{}
	for _, server := range job.Team.Servers {
		allowedIPs = append(allowedIPs, server.GlobalIP)
	}

	suffix, err := getExternalServiceSuffix()
	if err != nil {
		return &JobResult{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxBenchmarkTime)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		benchmarkerPath,
		fmt.Sprintf("-payment-url=https://%s", fmt.Sprintf("payment%s.isucon9q.catatsuy.org", suffix)),
		fmt.Sprintf("-shipment-url=https://%s", fmt.Sprintf("shipment%s.isucon9q.catatsuy.org", suffix)),
		fmt.Sprintf("-target-url=https://%s", target.GlobalIP),
		fmt.Sprintf("-allowed-ips=%s", strings.Join(allowedIPs, ",")),
		fmt.Sprintf("-data-dir=/home/isucon/isucari/initial-data"))

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// triming too long stderr
	stderrStr := stderr.String()
	if len(stderrStr) > maxStderrLength {
		stderrStr = stderrStr[:maxStderrLength]
	}

	return &JobResult{
		Stdout: stdout.String(),
		Stderr: stderrStr,
	}, err
}

func main() {

	apiEndpoint := flag.String("ep", apiEndpointDev, "API Endpoint")
	interval := flag.Duration("interval", defaultInterval, "Dequeuing interval second")
	benchmarkerPath := flag.String("benchmarker", defaultBenchmarkerPath, "Benchmarker path")
	flag.Parse()

	ticker := time.NewTicker(*interval)
	for range ticker.C {
		job, err := dequeue(*apiEndpoint)
		if err != nil {
			log.Println(err)
			continue
		}

		jobResult, err := runBenchmarker(*benchmarkerPath, job)
		if err != nil {
			log.Println(err)
		}

		log.Println(jobResult.Stderr)
		log.Println(jobResult.Stdout)
		if err := report(*apiEndpoint, job, jobResult); err != nil {
			log.Println(err)
		}
		log.Printf("job:%d reported", job.ID)
	}
}
