package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/podcctv/detective-chicken/internal/agent"
	"github.com/podcctv/detective-chicken/internal/model"
)

func main() {
	configDefault := "/etc/detective-chicken/agent.json"
	if dir, err := os.UserConfigDir(); err == nil && runtime.GOOS == "windows" {
		configDefault = filepath.Join(dir, "detective-chicken", "agent.json")
	}
	configPath := flag.String("config", configDefault, "agent config path")
	serverURL := flag.String("server", "http://127.0.0.1:8080", "Detective Chicken API base URL")
	token := flag.String("token", "", "one-time enrollment token")
	family := flag.String("family", "auto", "IP family to scan: auto, 4 or 6")
	scriptURL := flag.String("script-url", "https://IP.Check.Place", "IPQuality script URL")
	flag.Parse()
	if flag.NArg() == 0 {
		fatal("usage: detective-chicken-agent [flags] <enroll|heartbeat|scan>")
	}
	switch flag.Arg(0) {
	case "enroll":
		if *token == "" {
			fatal("--token is required")
		}
		cfg, err := agent.Enroll(*serverURL, *token, *configPath)
		if err != nil {
			fatal(err.Error())
		}
		fmt.Printf("registered agent %s for node %s\n", cfg.AgentID, cfg.NodeID)
	case "heartbeat":
		client := load(*configPath)
		directive, err := client.Heartbeat(nil)
		if err != nil {
			fatal(err.Error())
		}
		fmt.Printf("heartbeat accepted; quality interval %s\n", formatInterval(directive.ScanIntervalMinutes))
		hasScanCmd := false
		for _, cmd := range directive.Commands {
			if cmd.Type == "scan" {
				hasScanCmd = true
				break
			}
		}
		if directive.ScanDue || hasScanCmd {
			if err := scanAndUpload(client, *scriptURL, nil); err != nil {
				fatal(err.Error())
			}
		}

	case "scan":
		client := load(*configPath)
		families, err := requestedFamilies(*family)
		if err != nil {
			fatal(err.Error())
		}
		if err = scanAndUpload(client, *scriptURL, families); err != nil {
			fatal(err.Error())
		}
	default:
		fatal("unknown command: " + flag.Arg(0))
	}
}

type scanResult struct {
	family int
	report model.Report
	err    error
}

func requestedFamilies(value string) ([]int, error) {
	if value == "auto" {
		return nil, nil
	}
	family, err := strconv.Atoi(value)
	if err != nil || family != 4 && family != 6 {
		return nil, fmt.Errorf("family must be auto, 4 or 6")
	}
	return []int{family}, nil
}

func scanAndUpload(client *agent.Client, scriptURL string, families []int) error {
	if len(families) == 0 {
		var err error
		families, err = agent.AvailableFamilies()
		if err != nil {
			return fmt.Errorf("detect network families: %w", err)
		}
	}
	if len(families) == 0 {
		return fmt.Errorf("no usable IPv4 or IPv6 interface detected")
	}
	_, _ = client.Heartbeat(map[string]any{"state": "ready", "scan_state": "scanning", "scan_error": ""})
	fmt.Printf("quality scan started for IPv%v; usually 1-3 minutes, maximum 8 minutes\n", families)
	results := make(chan scanResult, len(families))
	for _, family := range families {
		go func(family int) {
			report, err := (agent.Adapter{ScriptURL: scriptURL}).Collect(context.Background(), family)
			results <- scanResult{family: family, report: report, err: err}
		}(family)
	}
	succeeded := 0
	var failures []string
	for range families {
		result := <-results
		if result.err == nil {
			result.err = client.Upload(result.report)
		}
		if result.err != nil {
			message := fmt.Sprintf("IPv%d: %s", result.family, shortError(result.err))
			failures = append(failures, message)
			fmt.Fprintln(os.Stderr, message)
			continue
		}
		succeeded++
		fmt.Printf("IPv%d report %s accepted\n", result.family, result.report.ReportID)
	}
	state := "ready"
	if len(failures) > 0 {
		state = "partial"
	}
	if succeeded == 0 {
		state = "failed"
	}
	_, _ = client.Heartbeat(map[string]any{"state": "ready", "scan_state": state, "scan_error": strings.Join(failures, "; ")})
	if succeeded == 0 {
		return fmt.Errorf("all quality scans failed: %s", strings.Join(failures, "; "))
	}
	return nil
}

func shortError(err error) string {
	message := strings.TrimSpace(err.Error())
	if len(message) > 500 {
		message = message[:500] + "..."
	}
	return message
}

func formatInterval(minutes int) string {
	if minutes <= 0 {
		minutes = 360
	}
	if minutes%1440 == 0 {
		return fmt.Sprintf("%dd", minutes/1440)
	}
	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}
	return fmt.Sprintf("%dmin", minutes)
}
func load(path string) *agent.Client {
	cfg, err := agent.LoadConfig(path)
	if err != nil {
		fatal(err.Error())
	}
	return &agent.Client{Config: cfg}
}
func fatal(message string) { fmt.Fprintln(os.Stderr, "detective-chicken-agent:", message); os.Exit(1) }
