package traxcli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/kamcpp/trax/pkg/cache"
	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/mq"
)

const (
	MaxTraxHistoryEntries = 100000
	TraxHistoryFileName   = ".traxcli_history"
	DefaultTraxCtrlURL    = "http://host.docker.internal:17202"
)

// Init initializes the context and logger for traxcli
func Init() context.Context {
	ctx := context.Background()
	common.SubComponent = "traxcli"
	common.InitLogger()
	if os.Getenv("SU_MODE") == "active" {
		common.L.Warn("!!! SU mode is active !!!", common.F(ctx)...)
	}
	cache.Init(ctx)
	mq.Init(ctx)
	return ctx
}

type TraxCli struct {
	baseURL       string
	ctx           context.Context
	readline      *readline.Instance
	historyPath   string
	lastInterrupt time.Time
	traceId       string
	client        *http.Client
}

// getTraxHistoryFilePath returns the path to the history file, preferring container-friendly locations
func getTraxHistoryFilePath() (string, error) {
	// Check for custom history directory via environment variable (container-friendly)
	if historyDir := os.Getenv("TRAXCLI_HISTORY_DIR"); historyDir != "" {
		return filepath.Join(historyDir, TraxHistoryFileName), nil
	}

	// Check if we're likely in a container - use /data for volume mounting
	if _, err := os.Stat("/data"); err == nil {
		return filepath.Join("/data", TraxHistoryFileName), nil
	}

	// Check if /tmp is writable (fallback for containers)
	if tmpDir := os.TempDir(); tmpDir != "" {
		return filepath.Join(tmpDir, TraxHistoryFileName), nil
	}

	// Finally, fallback to user home directory (for local development)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return filepath.Join(homeDir, TraxHistoryFileName), nil
}

// trimTraxHistoryFile ensures the history file doesn't exceed MaxTraxHistoryEntries
func trimTraxHistoryFile(historyPath string) error {
	file, err := os.Open(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to trim
		}
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read history file: %v", err)
	}

	// If we're under the limit, no need to trim
	if len(lines) <= MaxTraxHistoryEntries {
		return nil
	}

	// Keep only the most recent MaxTraxHistoryEntries
	linesToKeep := lines[len(lines)-MaxTraxHistoryEntries:]

	// Write the trimmed history back to file
	tmpFile := historyPath + ".tmp"
	outFile, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary history file: %v", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range linesToKeep {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to temporary history file: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush temporary history file: %v", err)
	}

	// Replace original file with trimmed version
	if err := os.Rename(tmpFile, historyPath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to replace history file: %v", err)
	}

	fmt.Printf("History file trimmed to %d entries\n", MaxTraxHistoryEntries)
	return nil
}

// RunInteractive starts the interactive trax CLI with optional trace ID
func RunInteractive(ctx context.Context, traceId string) {

	// Get history file path
	historyPath, err := getTraxHistoryFilePath()
	if err != nil {
		fmt.Printf("Warning: Could not get history file path: %v\n", err)
		historyPath = "/tmp/traxcli_history" // Fallback to temp directory
	}

	// Trim history file if it exceeds the limit
	if err := trimTraxHistoryFile(historyPath); err != nil {
		fmt.Printf("Warning: Could not trim history file: %v\n", err)
	}

	// Initialize readline with advanced command and argument completion
	completer := readline.NewPrefixCompleter(
		readline.PcItem("connect"),
		readline.PcItem("health",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
		),
		readline.PcItem("saga-template",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--id="),
		),
		readline.PcItem("saga-templates",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
		),
		readline.PcItem("saga-template-ids",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
		),
		readline.PcItem("saga-instance",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("saga-instances",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("saga-instance-ids",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("saga-step-instance",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("saga-step-instances",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("saga-step-instance-ids",
			readline.PcItem("--url="),
			readline.PcItem("--trace-id="),
			readline.PcItem("--cluster-id="),
		),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "traxcli> ",
		HistoryFile:  historyPath,
		AutoComplete: completer,
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	cli := &TraxCli{
		baseURL:     DefaultTraxCtrlURL,
		ctx:         ctx,
		readline:    rl,
		historyPath: historyPath,
		traceId:     traceId,
		client:      &http.Client{Timeout: 30 * time.Second},
	}

	fmt.Println("=== TraxCtrl CLI v1 ===")
	fmt.Println("Commands:")
	fmt.Println("  connect [url]                    - Connect to traxctrl server (defaults to http://host.docker.internal:17202)")
	fmt.Println("  health [--trace-id=ID]           - Check server health")
	fmt.Println("  saga-template --id=<id> [--trace-id=ID] - Get saga template by ID")
	fmt.Println("  saga-templates [--trace-id=ID]   - Get all saga templates")
	fmt.Println("  saga-template-ids [--trace-id=ID] - Get all saga template IDs only")
	fmt.Println("  saga-instance --id=<id> --cluster-id=<cluster_id> [--trace-id=ID] - Get saga instance by ID")
	fmt.Println("  saga-instances --cluster-id=<cluster_id> [--trace-id=ID] - Get all saga instances")
	fmt.Println("  saga-instance-ids --cluster-id=<cluster_id> [--trace-id=ID] - Get all saga instance IDs only")
	fmt.Println("  saga-step-instance --id=<id> --cluster-id=<cluster_id> [--trace-id=ID] - Get saga step instance by ID")
	fmt.Println("  saga-step-instances --cluster-id=<cluster_id> [--trace-id=ID] - Get all saga step instances")
	fmt.Println("  saga-step-instance-ids --cluster-id=<cluster_id> [--trace-id=ID] - Get all saga step instance IDs only")
	fmt.Println("  executor [options]               - Run a saga step executor")
	fmt.Println("  help                             - Show this help")
	fmt.Println("  exit                             - Exit CLI")
	fmt.Println("Features:")
	fmt.Println("  - Up/Down arrows for command history")
	fmt.Println("  - Tab completion for commands and arguments")
	fmt.Println("  - Left/Right arrows for cursor navigation")
	fmt.Printf("History file: %s\n", historyPath)
	if traceId != "" {
		fmt.Printf("Trace ID: %s\n", traceId)
	}
	fmt.Println()

	cli.runInteractiveLoop()
}

func (cli *TraxCli) runInteractiveLoop() {
	for {
		line, err := cli.readline.Readline()
		if err != nil {
			if err == io.EOF {
				break
			} else if err == readline.ErrInterrupt {
				// Ctrl+C pressed - check for double Ctrl+C
				now := time.Now()
				if !cli.lastInterrupt.IsZero() && now.Sub(cli.lastInterrupt) < 2*time.Second {
					// Double Ctrl+C within 2 seconds - exit
					fmt.Println("\nExiting...")
					break
				}
				// Single Ctrl+C - clear line and continue
				cli.lastInterrupt = now
				fmt.Println("Press Ctrl+C again to exit, or type 'exit'")
				continue
			} else {
				fmt.Printf("Error reading input: %v\n", err)
				break
			}
		}

		// Reset interrupt timer on successful input
		cli.lastInterrupt = time.Time{}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := cli.processCommand(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		if line == "exit" {
			break
		}
	}

	// Save history explicitly before exiting
	if err := cli.readline.SaveHistory(cli.historyPath); err != nil {
		fmt.Printf("Warning: Failed to save history: %v\n", err)
	}
}

func (cli *TraxCli) processCommand(line string) error {
	parts := cli.parseCommand(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "connect":
		return cli.cmdConnect(args)
	case "health":
		return cli.cmdHealth(args)
	case "saga-template":
		return cli.cmdSagaTemplate(args)
	case "saga-templates":
		return cli.cmdSagaTemplates(args)
	case "saga-template-ids":
		return cli.cmdSagaTemplateIds(args)
	case "saga-instance":
		return cli.cmdSagaInstance(args)
	case "saga-instances":
		return cli.cmdSagaInstances(args)
	case "saga-instance-ids":
		return cli.cmdSagaInstanceIds(args)
	case "saga-step-instance":
		return cli.cmdSagaStepInstance(args)
	case "saga-step-instances":
		return cli.cmdSagaStepInstances(args)
	case "saga-step-instance-ids":
		return cli.cmdSagaStepInstanceIds(args)
	case "saga-template-update":
		return cli.cmdSagaTemplateUpdate(args)
	case "saga-template-delete":
		return cli.cmdSagaTemplateDelete(args)
	case "saga-step-template-update":
		return cli.cmdSagaStepTemplateUpdate(args)
	case "saga-step-template-delete":
		return cli.cmdSagaStepTemplateDelete(args)
	case "executor":
		return cli.cmdExecutor(args)
	case "help":
		return cli.cmdHelp(args)
	case "exit":
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func (cli *TraxCli) parseCommand(line string) []string {
	re := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	matches := re.FindAllString(line, -1)

	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if strings.HasPrefix(match, `"`) && strings.HasSuffix(match, `"`) {
			result = append(result, match[1:len(match)-1])
		} else if strings.HasPrefix(match, `'`) && strings.HasSuffix(match, `'`) {
			result = append(result, match[1:len(match)-1])
		} else {
			result = append(result, match)
		}
	}
	return result
}

// extractFlags extracts --url, --trace-id, --json, --verbose, --id, and --cluster-id from args
func (cli *TraxCli) extractFlags(args []string) (url, traceId, id, clusterId string, jsonOutput, verbose bool, filteredArgs []string) {
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--url=") {
			url = strings.TrimPrefix(args[i], "--url=")
		} else if strings.HasPrefix(args[i], "--trace-id=") {
			traceId = strings.TrimPrefix(args[i], "--trace-id=")
		} else if strings.HasPrefix(args[i], "--id=") {
			id = strings.TrimPrefix(args[i], "--id=")
		} else if strings.HasPrefix(args[i], "--cluster-id=") {
			clusterId = strings.TrimPrefix(args[i], "--cluster-id=")
		} else if args[i] == "--json" {
			jsonOutput = true
		} else if args[i] == "--verbose" || args[i] == "-v" {
			verbose = true
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if url == "" {
		url = cli.baseURL
	}

	if traceId == "" {
		traceId = cli.traceId
	}

	return url, traceId, id, clusterId, jsonOutput, verbose, filteredArgs
}

func (cli *TraxCli) cmdConnect(args []string) error {
	var url string

	if len(args) == 0 {
		url = DefaultTraxCtrlURL
	} else if len(args) == 1 {
		url = args[0]
	} else {
		return fmt.Errorf("usage: connect [url] (defaults to %s)", DefaultTraxCtrlURL)
	}

	// Test the connection
	fmt.Printf("Connecting to %s...\n", url)

	req, err := http.NewRequestWithContext(cli.ctx, "GET", url+"/api/v1/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := cli.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	cli.baseURL = url
	fmt.Println("Connected successfully!")
	return nil
}

func (cli *TraxCli) makeRequest(method, endpoint string, body []byte, traceId string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(cli.ctx, method, cli.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if traceId != "" {
		req.Header.Set("x-trace-id", traceId)
	}

	return cli.client.Do(req)
}

func (cli *TraxCli) cmdHealth(args []string) error {
	url, traceId, _, _, _, verbose, _ := cli.extractFlags(args)

	if verbose {
		fmt.Printf("Checking health at %s...\n", url)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	resp, err := cli.makeRequest("GET", "/api/v1/health", nil, traceId)
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	fmt.Printf("Response: %s\n", string(body))
	return nil
}

func (cli *TraxCli) cmdSagaTemplate(args []string) error {
	url, traceId, templateId, _, jsonOutput, verbose, _ := cli.extractFlags(args)

	if templateId == "" {
		return fmt.Errorf("--id is required")
	}

	if verbose {
		fmt.Printf("Getting saga template: %s\n", templateId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	endpoint := fmt.Sprintf("/api/v1/saga-templates/%s", templateId)
	resp, err := cli.makeRequest("POST", endpoint, nil, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.formatSagaTemplateDetails(body)
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaTemplates(args []string) error {
	url, traceId, _, _, jsonOutput, verbose, _ := cli.extractFlags(args)

	if verbose {
		fmt.Println("Getting all saga templates...")
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	resp, err := cli.makeRequest("POST", "/api/v1/saga-templates/list", nil, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.formatSagaTemplateTable(body)
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaTemplateIds(args []string) error {
	url, traceId, _, _, jsonOutput, verbose, _ := cli.extractFlags(args)

	if verbose {
		fmt.Println("Getting saga template IDs...")
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	resp, err := cli.makeRequest("POST", "/api/v1/saga-templates/list/ids", nil, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.formatSagaTemplateIds(body)
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

// cmdSagaTemplateUpdate updates a saga template via the traxctrl REST API.
// Usage: saga-template-update --id <template_id> --display-name "New Name" [--description "..."] [--url <url>]
func (cli *TraxCli) cmdSagaTemplateUpdate(args []string) error {
	url, traceId, templateId, _, _, verbose, filteredArgs := cli.extractFlags(args)

	if templateId == "" {
		return fmt.Errorf("--id is required")
	}

	// Parse update-specific flags from filteredArgs
	displayName, description := "", ""
	for i := 0; i < len(filteredArgs); i++ {
		if strings.HasPrefix(filteredArgs[i], "--display-name=") {
			displayName = strings.TrimPrefix(filteredArgs[i], "--display-name=")
		} else if strings.HasPrefix(filteredArgs[i], "--description=") {
			description = strings.TrimPrefix(filteredArgs[i], "--description=")
		}
	}

	payload := map[string]interface{}{}
	if displayName != "" {
		payload["display_name"] = displayName
	}
	if description != "" {
		payload["description"] = description
	}

	if len(payload) == 0 {
		return fmt.Errorf("at least one of --display-name or --description is required")
	}

	if verbose {
		fmt.Printf("Updating saga template: %s\n", templateId)
	}

	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	payloadBytes, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("/api/v1/saga-templates/%s", templateId)
	resp, err := cli.makeRequest("PUT", endpoint, payloadBytes, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ Saga template '%s' updated\n", templateId)
	} else {
		fmt.Printf("✗ Update failed (%d): %s\n", resp.StatusCode, string(body))
	}
	return nil
}

// cmdSagaTemplateDelete deletes a saga template and its step templates via the traxctrl REST API.
// Usage: saga-template-delete --id <template_id> [--url <url>]
func (cli *TraxCli) cmdSagaTemplateDelete(args []string) error {
	url, traceId, templateId, _, _, verbose, _ := cli.extractFlags(args)

	if templateId == "" {
		return fmt.Errorf("--id is required")
	}

	if verbose {
		fmt.Printf("Deleting saga template: %s\n", templateId)
	}

	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	endpoint := fmt.Sprintf("/api/v1/saga-templates/%s", templateId)
	resp, err := cli.makeRequest("DELETE", endpoint, nil, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ Saga template '%s' deleted\n", templateId)
	} else {
		fmt.Printf("✗ Delete failed (%d): %s\n", resp.StatusCode, string(body))
	}
	return nil
}

// cmdSagaStepTemplateUpdate updates a saga step template via the traxctrl REST API.
// Usage: saga-step-template-update --id <step_template_id> --display-name "New Name" [--url <url>]
func (cli *TraxCli) cmdSagaStepTemplateUpdate(args []string) error {
	url, traceId, stepTemplateId, _, _, verbose, filteredArgs := cli.extractFlags(args)

	if stepTemplateId == "" {
		return fmt.Errorf("--id is required")
	}

	displayName, description, sagaTemplateId := "", "", ""
	for i := 0; i < len(filteredArgs); i++ {
		if strings.HasPrefix(filteredArgs[i], "--display-name=") {
			displayName = strings.TrimPrefix(filteredArgs[i], "--display-name=")
		} else if strings.HasPrefix(filteredArgs[i], "--description=") {
			description = strings.TrimPrefix(filteredArgs[i], "--description=")
		} else if strings.HasPrefix(filteredArgs[i], "--saga-template-id=") {
			sagaTemplateId = strings.TrimPrefix(filteredArgs[i], "--saga-template-id=")
		}
	}

	payload := map[string]interface{}{}
	if displayName != "" {
		payload["display_name"] = displayName
	}
	if description != "" {
		payload["description"] = description
	}
	if sagaTemplateId != "" {
		payload["saga_template_id"] = sagaTemplateId
	}

	if len(payload) == 0 {
		return fmt.Errorf("at least one of --display-name, --description, or --saga-template-id is required")
	}
	// saga_template_id is required by the API
	if sagaTemplateId == "" {
		return fmt.Errorf("--saga-template-id is required for step template update")
	}

	if verbose {
		fmt.Printf("Updating saga step template: %s\n", stepTemplateId)
	}

	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	payloadBytes, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("/api/v1/saga-step-templates/%s", stepTemplateId)
	resp, err := cli.makeRequest("PUT", endpoint, payloadBytes, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ Saga step template '%s' updated\n", stepTemplateId)
	} else {
		fmt.Printf("✗ Update failed (%d): %s\n", resp.StatusCode, string(body))
	}
	return nil
}

// cmdSagaStepTemplateDelete deletes a saga step template via the traxctrl REST API.
// Usage: saga-step-template-delete --id <step_template_id> [--url <url>]
func (cli *TraxCli) cmdSagaStepTemplateDelete(args []string) error {
	url, traceId, stepTemplateId, _, _, verbose, _ := cli.extractFlags(args)

	if stepTemplateId == "" {
		return fmt.Errorf("--id is required")
	}

	if verbose {
		fmt.Printf("Deleting saga step template: %s\n", stepTemplateId)
	}

	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	endpoint := fmt.Sprintf("/api/v1/saga-step-templates/%s", stepTemplateId)
	resp, err := cli.makeRequest("DELETE", endpoint, nil, traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ Saga step template '%s' deleted\n", stepTemplateId)
	} else {
		fmt.Printf("✗ Delete failed (%d): %s\n", resp.StatusCode, string(body))
	}
	return nil
}

func (cli *TraxCli) prettyPrintJSON(data []byte) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Printf("Response: %s\n", string(data))
		return
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Printf("Response: %s\n", string(data))
		return
	}

	fmt.Printf("Response:\n%s\n", string(prettyJSON))
}

func (cli *TraxCli) printJSONOnly(data []byte) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Printf("%s\n", string(data))
		return
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Printf("%s\n", string(data))
		return
	}

	fmt.Printf("%s\n", string(prettyJSON))
}

type SagaTemplate struct {
	TemplateId          string             `json:"template_id"`
	DisplayName         string             `json:"display_name"`
	Description         string             `json:"description"`
	Tags                []string           `json:"tags"`
	SagaStepTemplateIds []string           `json:"saga_step_template_ids"`
	SagaStepTemplates   []SagaStepTemplate `json:"saga_step_templates"`
}

type SagaStepTemplate struct {
	TemplateId     string   `json:"template_id"`
	DisplayName    string   `json:"display_name"`
	Description    string   `json:"description"`
	SagaTemplateId string   `json:"saga_template_id"`
	Tags           []string `json:"tags"`
}

func (cli *TraxCli) formatSagaTemplateTable(data []byte) {
	var response struct {
		SagaTemplates []SagaTemplate `json:"saga_templates"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(response.SagaTemplates) == 0 {
		fmt.Println("No saga templates found.")
		return
	}

	fmt.Printf("%-40s %-30s %-50s %-15s\n", "TEMPLATE ID", "DISPLAY NAME", "DESCRIPTION", "STEP COUNT")
	fmt.Printf("%-40s %-30s %-50s %-15s\n", strings.Repeat("-", 40), strings.Repeat("-", 30), strings.Repeat("-", 50), strings.Repeat("-", 15))

	for _, template := range response.SagaTemplates {
		description := template.Description
		if len(description) > 47 {
			description = description[:47] + "..."
		}
		fmt.Printf("%-40s %-30s %-50s %-15d\n",
			template.TemplateId,
			template.DisplayName,
			description,
			len(template.SagaStepTemplates))
	}
}

func (cli *TraxCli) formatSagaTemplateDetails(data []byte) {
	var template SagaTemplate

	if err := json.Unmarshal(data, &template); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("Template ID:    %s\n", template.TemplateId)
	fmt.Printf("Display Name:   %s\n", template.DisplayName)
	fmt.Printf("Description:    %s\n", template.Description)
	fmt.Printf("Tags:           %s\n", strings.Join(template.Tags, ", "))
	fmt.Printf("Step Count:     %d\n", len(template.SagaStepTemplates))

	if len(template.SagaStepTemplates) > 0 {
		fmt.Printf("\nSteps:\n")
		fmt.Printf("  %-35s %-25s %-40s\n", "STEP ID", "DISPLAY NAME", "DESCRIPTION")
		fmt.Printf("  %-35s %-25s %-40s\n", strings.Repeat("-", 35), strings.Repeat("-", 25), strings.Repeat("-", 40))

		for _, step := range template.SagaStepTemplates {
			description := step.Description
			if len(description) > 37 {
				description = description[:37] + "..."
			}
			fmt.Printf("  %-35s %-25s %-40s\n", step.TemplateId, step.DisplayName, description)
		}
	}
}

func (cli *TraxCli) formatSagaTemplateIds(data []byte) {
	var response struct {
		SagaTemplateIds []string `json:"saga_template_ids"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(response.SagaTemplateIds) == 0 {
		fmt.Println("No saga template IDs found.")
		return
	}

	fmt.Println("Saga Template IDs:")
	sort.Strings(response.SagaTemplateIds)
	for i, id := range response.SagaTemplateIds {
		fmt.Printf("%d. %s\n", i+1, id)
	}
}

func (cli *TraxCli) cmdSagaInstance(args []string) error {
	url, traceId, instanceId, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if instanceId == "" {
		return fmt.Errorf("--id is required")
	}

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting saga instance: %s (cluster: %s)\n", instanceId, clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	endpoint := fmt.Sprintf("/api/v1/saga-instances/%s", instanceId)
	resp, err := cli.makeRequest("POST", endpoint, []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaInstances(args []string) error {
	url, traceId, _, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting all saga instances for cluster: %s\n", clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	resp, err := cli.makeRequest("POST", "/api/v1/saga-instances/list", []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaInstanceIds(args []string) error {
	url, traceId, _, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting saga instance IDs for cluster: %s\n", clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	resp, err := cli.makeRequest("POST", "/api/v1/saga-instances/list/ids", []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaStepInstance(args []string) error {
	url, traceId, instanceId, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if instanceId == "" {
		return fmt.Errorf("--id is required")
	}

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting saga step instance: %s (cluster: %s)\n", instanceId, clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	endpoint := fmt.Sprintf("/api/v1/saga-step-instances/%s", instanceId)
	resp, err := cli.makeRequest("POST", endpoint, []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaStepInstances(args []string) error {
	url, traceId, _, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting all saga step instances for cluster: %s\n", clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	resp, err := cli.makeRequest("POST", "/api/v1/saga-step-instances/list", []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdSagaStepInstanceIds(args []string) error {
	url, traceId, _, clusterId, jsonOutput, verbose, _ := cli.extractFlags(args)

	if clusterId == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	if verbose {
		fmt.Printf("Getting saga step instance IDs for cluster: %s\n", clusterId)
		if traceId != "" {
			fmt.Printf("Using trace-id: %s\n", traceId)
		}
	}

	// Temporarily override baseURL if different URL provided
	originalURL := cli.baseURL
	cli.baseURL = url
	defer func() { cli.baseURL = originalURL }()

	requestBody := fmt.Sprintf(`{"cluster_id":"%s"}`, clusterId)
	resp, err := cli.makeRequest("POST", "/api/v1/saga-step-instances/list/ids", []byte(requestBody), traceId)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if verbose {
		fmt.Printf("Status: %d\n", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		if jsonOutput {
			if verbose {
				cli.prettyPrintJSON(body)
			} else {
				cli.printJSONOnly(body)
			}
		} else {
			cli.prettyPrintJSON(body) // For now, always print JSON until we implement formatting
		}
	} else {
		if verbose {
			fmt.Printf("Response: %s\n", string(body))
		} else {
			fmt.Printf("%s\n", string(body))
		}
	}
	return nil
}

func (cli *TraxCli) cmdExecutor(args []string) error {
	// Parse executor arguments
	cfg := &ExecutorConfig{}

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--trax-cluster-id=") {
			cfg.TraxClusterId = strings.TrimPrefix(args[i], "--trax-cluster-id=")
		} else if strings.HasPrefix(args[i], "--rabbitmq-url=") {
			cfg.RabbitmqURL = strings.TrimPrefix(args[i], "--rabbitmq-url=")
		} else if strings.HasPrefix(args[i], "--saga-template-id=") {
			cfg.SagaTemplateId = strings.TrimPrefix(args[i], "--saga-template-id=")
		} else if strings.HasPrefix(args[i], "--saga-step-template-id=") {
			cfg.SagaStepTemplateId = strings.TrimPrefix(args[i], "--saga-step-template-id=")
		} else if strings.HasPrefix(args[i], "--exec-sim-status=") {
			cfg.ExecSimStatus = strings.TrimPrefix(args[i], "--exec-sim-status=")
		} else if strings.HasPrefix(args[i], "--exec-sim-delay=") {
			cfg.ExecSimDelay = strings.TrimPrefix(args[i], "--exec-sim-delay=")
		} else if strings.HasPrefix(args[i], "--exec-sim-error=") {
			cfg.ExecSimError = strings.TrimPrefix(args[i], "--exec-sim-error=")
		} else if strings.HasPrefix(args[i], "--exec-sim-result=") {
			cfg.ExecSimResult = strings.TrimPrefix(args[i], "--exec-sim-result=")
		} else if strings.HasPrefix(args[i], "--mq-event-pub-node=") {
			cfg.MqEventPubNode = strings.TrimPrefix(args[i], "--mq-event-pub-node=")
		} else if strings.HasPrefix(args[i], "--exec-shell=") {
			cfg.ExecShell = strings.TrimPrefix(args[i], "--exec-shell=")
		} else if strings.HasPrefix(args[i], "--idempotency-storage-backend=") {
			cfg.IdempotencyBackend = strings.TrimPrefix(args[i], "--idempotency-storage-backend=")
		} else if strings.HasPrefix(args[i], "--redis-url=") {
			cfg.RedisURL = strings.TrimPrefix(args[i], "--redis-url=")
		} else if strings.HasPrefix(args[i], "--pgsql-url=") {
			cfg.PgsqlURL = strings.TrimPrefix(args[i], "--pgsql-url=")
		}
	}

	// Validate required fields
	if cfg.TraxClusterId == "" {
		return fmt.Errorf("--trax-cluster-id is required")
	}
	if cfg.RabbitmqURL == "" {
		return fmt.Errorf("--rabbitmq-url is required")
	}
	if cfg.SagaTemplateId == "" {
		return fmt.Errorf("--saga-template-id is required")
	}
	if cfg.SagaStepTemplateId == "" {
		return fmt.Errorf("--saga-step-template-id is required")
	}

	// Run executor
	return RunExecutor(cli.ctx, cfg)
}

func (cli *TraxCli) cmdHelp(_ []string) error {
	fmt.Println("Available commands:")
	fmt.Println("  connect [url]                           - Connect to traxctrl server (defaults to http://host.docker.internal:17202)")
	fmt.Println("  health [--url=URL] [--trace-id=ID] [-v|--verbose] - Check server health")
	fmt.Println("  saga-template --id=<id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get saga template by ID")
	fmt.Println("  saga-templates [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga templates")
	fmt.Println("  saga-template-ids [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga template IDs only")
	fmt.Println("  saga-instance --id=<id> --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get saga instance by ID")
	fmt.Println("  saga-instances --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga instances")
	fmt.Println("  saga-instance-ids --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga instance IDs only")
	fmt.Println("  saga-step-instance --id=<id> --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get saga step instance by ID")
	fmt.Println("  saga-step-instances --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga step instances")
	fmt.Println("  saga-step-instance-ids --cluster-id=<cluster_id> [--url=URL] [--trace-id=ID] [--json] [-v|--verbose] - Get all saga step instance IDs only")
	fmt.Println("  executor [options]                      - Run a saga step executor (type 'executor --help' for details)")
	fmt.Println("  help                                    - Show this help")
	fmt.Println("  exit                                    - Exit CLI")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --json                                  - Output in JSON format (default: table format)")
	fmt.Println("  -v, --verbose                           - Show verbose output (status, trace info)")
	fmt.Println("                                          - For JSON output: shows 'Status:', 'Response:' headers")
	fmt.Println("                                          - For table output: shows connection/request info")
	fmt.Println()
	fmt.Println("URL options:")
	fmt.Println("  --url=http://server:port  - Override default server URL")
	fmt.Println("  (uses connected server if not specified)")
	fmt.Println()
	fmt.Println("Trace ID options:")
	fmt.Println("  --trace-id=myid  - Use custom trace ID (sent as x-trace-id header)")
	fmt.Println("  (uses default context if not specified)")
	fmt.Println()
	fmt.Println("Terminal features:")
	fmt.Println("  Up/Down arrows   - Navigate command history")
	fmt.Println("  Left/Right arrows - Move cursor in current line")
	fmt.Println("  Tab              - Auto-complete commands and arguments")
	fmt.Println("  Ctrl+C           - Clear line (double Ctrl+C to exit)")
	fmt.Println("  Ctrl+D           - Exit CLI")
	fmt.Println("  History file     - Auto-located (max 100K entries)")
	fmt.Println()
	fmt.Println("History file locations (in order of preference):")
	fmt.Println("  $TRAXCLI_HISTORY_DIR/.traxcli_history - Custom directory")
	fmt.Println("  /data/.traxcli_history                - Container volume mount")
	fmt.Println("  /tmp/.traxcli_history                 - Container fallback")
	fmt.Println("  ~/.traxcli_history                    - Local development")
	fmt.Println()
	fmt.Println("Container usage:")
	fmt.Println("  docker run -v /host/path:/data ... traxcli")
	fmt.Println("  docker run -e TRAXCLI_HISTORY_DIR=/custom ... traxcli")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  connect http://localhost:17202")
	fmt.Println("  health --trace-id=test123")
	fmt.Println("  saga-template --id=new_account_under_participant")
	fmt.Println("  saga-templates --url=http://prod-server:17202")
	fmt.Println("  saga-template-ids")
	return nil
}

// RunNonInteractive executes a single traxcli command non-interactively
func RunNonInteractive(ctx context.Context, command string, args []string, traceId string) {
	// Create a CLI instance for non-interactive use
	cli := &TraxCli{
		baseURL: DefaultTraxCtrlURL,
		ctx:     ctx,
		traceId: traceId,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	// Execute the command directly
	commandLine := command
	if len(args) > 0 {
		commandLine = command + " " + strings.Join(args, " ")
	}

	fmt.Printf("Executing: %s\n", commandLine)
	if traceId != "" {
		fmt.Printf("Trace ID: %s\n", traceId)
	}

	if err := cli.processCommand(commandLine); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
