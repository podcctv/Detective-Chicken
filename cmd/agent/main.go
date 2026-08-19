package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/podcctv/detective-chicken/internal/agent"
)

func main() {
	configDefault := "/etc/detective-chicken/agent.json"
	if dir, err := os.UserConfigDir(); err == nil && runtime.GOOS == "windows" {
		configDefault = filepath.Join(dir, "detective-chicken", "agent.json")
	}
	configPath := flag.String("config", configDefault, "agent config path")
	serverURL := flag.String("server", "http://127.0.0.1:8080", "Detective Chicken API base URL")
	token := flag.String("token", "", "one-time enrollment token")
	family := flag.Int("family", 4, "IP family to scan: 4 or 6")
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
		if err := client.Heartbeat(); err != nil {
			fatal(err.Error())
		}
		fmt.Println("heartbeat accepted")
	case "scan":
		client := load(*configPath)
		report, err := (agent.Adapter{ScriptURL: *scriptURL}).Collect(context.Background(), *family)
		if err != nil {
			fatal(err.Error())
		}
		if err = client.Upload(report); err != nil {
			fatal(err.Error())
		}
		fmt.Printf("report %s accepted\n", report.ReportID)
	default:
		fatal("unknown command: " + flag.Arg(0))
	}
}
func load(path string) *agent.Client {
	cfg, err := agent.LoadConfig(path)
	if err != nil {
		fatal(err.Error())
	}
	return &agent.Client{Config: cfg}
}
func fatal(message string) { fmt.Fprintln(os.Stderr, "detective-chicken-agent:", message); os.Exit(1) }
