package traxcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/xshyft/trax/pkg/cache"
	"github.com/xshyft/trax/pkg/common"
	"github.com/xshyft/trax/pkg/mq"
	mqcommon "github.com/xshyft/trax/pkg/mq/common"
	"github.com/xshyft/trax/pkg/trax"
)

type ExecutorConfig struct {
	TraxClusterId      string
	SagaTemplateId     string
	SagaStepTemplateId string
	MqEventPubNode     string

	RabbitmqURL string
	RedisURL    string
	PgsqlURL    string

	ExecSimStatus string
	ExecSimDelay  string
	ExecSimError  string
	ExecSimResult string

	CompSimStatus string
	CompSimDelay  string
	CompSimError  string
	CompSimResult string

	ExecShell          string
	ExecShellPreDelay  string
	ExecShellPostDelay string

	CompShell          string
	CompShellPreDelay  string
	CompShellPostDelay string

	// Sub-saga spawning
	SubSagaTemplateId string
	TraxCtrlURL       string

	IdempotencyBackend string
}

func RunExecutor(ctx context.Context, cfg *ExecutorConfig) error {
	common.SubComponent = "traxcli.executor"
	common.InitLogger()

	if os.Getenv("SU_MODE") == "active" {
		common.L.Warn("!!! SU mode is active !!!", common.F(ctx)...)
	}

	// Validate configuration
	if err := validateExecutorConfig(cfg); err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cache.RedisAddr = cfg.RedisURL
	cache.Init(ctx)

	// Initialize RabbitMQ connection
	mqcommon.RabbitMQURL = cfg.RabbitmqURL
	mq.Init(ctx)

	// Create idempotency service based on configuration
	idempotentService, err := createIdempotentService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create idempotent service: %w", err)
	}

	// Create and run executor
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mqClient := trax.NewRabbitMQClient()

	// Build executor options
	var execOpts []trax.ExecutorOption

	// If sub-saga mode is configured, create a saga submitter
	if cfg.SubSagaTemplateId != "" && cfg.TraxCtrlURL != "" {
		submitterID := fmt.Sprintf("traxcli-sub-saga-%s-%s", cfg.SagaTemplateId, cfg.SagaStepTemplateId)
		sagaSubmitter := trax.NewDefaultSagaSubmitter(submitterID, mqClient)
		go sagaSubmitter.StartAnnouncement(ctx)

		common.L.Info(fmt.Sprintf(
			"Sub-saga mode: waiting for submitter '%s' to be ready (will spawn '%s')",
			submitterID, cfg.SubSagaTemplateId))

		if err := sagaSubmitter.WaitUntilReadyToAcceptSagaSubmissionRequests(ctx); err != nil {
			return fmt.Errorf("saga submitter failed to become ready: %w", err)
		}
		common.L.Info(fmt.Sprintf("Sub-saga submitter '%s' is ready", submitterID))

		execOpts = append(execOpts, trax.WithExecutorSagaSubmitter(sagaSubmitter))
		execOpts = append(execOpts, trax.WithExecutorTraxCtrlURL(cfg.TraxCtrlURL))
	}

	executor := trax.NewExecutor(
		mqClient,
		cfg.TraxClusterId,
		cfg.SagaTemplateId,
		cfg.SagaStepTemplateId,
		idempotentService,
		execOpts...,
	)

	common.L.Info(fmt.Sprintf(
		"Starting executor for saga %s, step %s on cluster %s",
		cfg.SagaTemplateId, cfg.SagaStepTemplateId, cfg.TraxClusterId))

	// Run executor in goroutine
	go func() {
		if err := executor.Run(ctx); err != nil {
			common.L.Error(fmt.Sprintf("Executor error: %v", err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	common.L.Info("Shutting down executor...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutines time to cleanup

	return nil
}

func validateExecutorConfig(cfg *ExecutorConfig) error {
	// Check that either simulation or shell mode is configured
	hasSimMode := cfg.ExecSimStatus != ""
	hasShellMode := cfg.ExecShell != ""

	if !hasSimMode && !hasShellMode {
		return errors.New("either simulation mode (--exec-sim-status) or shell execution mode (--exec-shell) must be specified")
	}

	if hasSimMode && hasShellMode {
		return errors.New("cannot specify both simulation mode and shell execution mode")
	}

	// Validate simulation mode configuration
	if hasSimMode {
		validStatuses := map[string]bool{"ok": true, "error": true, "noreturn": true, "sub-saga": true}
		if !validStatuses[cfg.ExecSimStatus] {
			return errors.New("--exec-sim-status must be one of: ok, error, noreturn, sub-saga")
		}

		// Sub-saga mode validation
		if cfg.ExecSimStatus == "sub-saga" {
			if cfg.SubSagaTemplateId == "" {
				return errors.New("--sub-saga-template-id is required when --exec-sim-status is sub-saga")
			}
			if cfg.TraxCtrlURL == "" {
				return errors.New("--traxctrl-url is required when --exec-sim-status is sub-saga")
			}
		}

		// Set default delay if not specified
		if cfg.ExecSimDelay == "" {
			cfg.ExecSimDelay = "0s"
		}

		// Validate delay duration format
		if _, err := time.ParseDuration(cfg.ExecSimDelay); err != nil {
			return fmt.Errorf("invalid --exec-sim-delay format: %w", err)
		}

		// Validate status-specific requirements
		if cfg.ExecSimStatus == "ok" && cfg.ExecSimResult == "" {
			return errors.New("--exec-sim-result is required when --exec-sim-status is ok")
		}

		if cfg.ExecSimStatus == "error" && cfg.ExecSimError == "" && cfg.ExecSimStatus != "sub-saga" {
			return errors.New("--exec-sim-error is required when --exec-sim-status is error")
		}

		// Validate JSON format
		if cfg.ExecSimResult != "" {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(cfg.ExecSimResult), &result); err != nil {
				return fmt.Errorf("invalid --exec-sim-result JSON format: %w", err)
			}
		}

		if cfg.ExecSimError != "" {
			var errorObj map[string]interface{}
			if err := json.Unmarshal([]byte(cfg.ExecSimError), &errorObj); err != nil {
				return fmt.Errorf("invalid --exec-sim-error JSON format: %w", err)
			}
		}

		// Validate compensation configuration (either simulation or shell)
		// Sub-saga executors don't need explicit compensation configuration because
		// compensation of a spawn step involves tracking child saga state, not running
		// a local compensation handler.
		hasCompSimMode := cfg.CompSimStatus != ""
		hasCompShellMode := cfg.CompShell != ""

		if cfg.ExecSimStatus != "sub-saga" && !hasCompSimMode && !hasCompShellMode {
			return errors.New("either compensation simulation mode (--comp-sim-status) or compensation shell mode (--comp-shell) must be specified")
		}

		if hasCompSimMode && hasCompShellMode {
			return errors.New("cannot specify both compensation simulation mode and compensation shell mode")
		}

		// Validate compensation simulation configuration
		if hasCompSimMode {
			if cfg.CompSimStatus != "ok" && cfg.CompSimStatus != "error" && cfg.CompSimStatus != "noreturn" {
				return errors.New("--comp-sim-status must be one of: ok, error, noreturn")
			}

			// Validate status-specific requirements
			if cfg.CompSimStatus == "error" && cfg.CompSimError == "" {
				return errors.New("--comp-sim-error is required when --comp-sim-status is error")
			}

			// Validate JSON format
			if cfg.CompSimResult != "" {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(cfg.CompSimResult), &result); err != nil {
					return fmt.Errorf("invalid --comp-sim-result JSON format: %w", err)
				}
			}

			if cfg.CompSimError != "" {
				var errorObj map[string]interface{}
				if err := json.Unmarshal([]byte(cfg.CompSimError), &errorObj); err != nil {
					return fmt.Errorf("invalid --comp-sim-error JSON format: %w", err)
				}
			}

			// Set compensation defaults if not specified
			if cfg.CompSimDelay == "" {
				cfg.CompSimDelay = "0s"
			}

			// Validate comp delay duration format
			if _, err := time.ParseDuration(cfg.CompSimDelay); err != nil {
				return fmt.Errorf("invalid --comp-sim-delay format: %w", err)
			}
		}
	}

	// Set shell delay defaults and validate
	if cfg.ExecShellPreDelay == "" {
		cfg.ExecShellPreDelay = "0s"
	}
	if cfg.ExecShellPostDelay == "" {
		cfg.ExecShellPostDelay = "0s"
	}
	if cfg.CompShellPreDelay == "" {
		cfg.CompShellPreDelay = "0s"
	}
	if cfg.CompShellPostDelay == "" {
		cfg.CompShellPostDelay = "0s"
	}

	// Validate shell delay duration formats
	if _, err := time.ParseDuration(cfg.ExecShellPreDelay); err != nil {
		return fmt.Errorf("invalid --exec-shell-predelay format: %w", err)
	}
	if _, err := time.ParseDuration(cfg.ExecShellPostDelay); err != nil {
		return fmt.Errorf("invalid --exec-shell-postdelay format: %w", err)
	}
	if _, err := time.ParseDuration(cfg.CompShellPreDelay); err != nil {
		return fmt.Errorf("invalid --comp-shell-predelay format: %w", err)
	}
	if _, err := time.ParseDuration(cfg.CompShellPostDelay); err != nil {
		return fmt.Errorf("invalid --comp-shell-postdelay format: %w", err)
	}

	// Validate idempotency backend configuration - REQUIRED for correct saga execution
	// Without idempotency, re-delivered messages will cause duplicate step executions
	if cfg.IdempotencyBackend == "" {
		panic("CRITICAL: --idempotency-storage-backend is REQUIRED. " +
			"Without idempotency tracking, message re-deliveries will cause duplicate step executions. " +
			"Use one of: inmem (for testing), redis, or pgsql")
	}

	if cfg.IdempotencyBackend != "inmem" && cfg.IdempotencyBackend != "redis" && cfg.IdempotencyBackend != "pgsql" {
		return errors.New("--idempotency-storage-backend must be one of: inmem, redis, pgsql")
	}

	if cfg.IdempotencyBackend == "redis" && cfg.RedisURL == "" {
		return errors.New("--redis-addr is required when --idempotency-storage-backend is redis")
	}

	if cfg.IdempotencyBackend == "pgsql" && cfg.PgsqlURL == "" {
		return errors.New("--pgsql-url is required when --idempotency-storage-backend is pgsql")
	}

	return nil
}

func createIdempotentService(cfg *ExecutorConfig) (trax.IdempotentService, error) {
	return newIdempotentService(cfg), nil
}

// idempotentService implements both simulation and shell-based execution/compensation
// It can handle any combination: exec-sim with comp-sim, exec-sim with comp-shell,
// exec-shell with comp-sim, or exec-shell with comp-shell
type idempotentService struct {
	cfg                 *ExecutorConfig
	executionResults    map[string]*trax.IdempotentServiceExecutionResult
	compensationResults map[string]*trax.IdempotentServiceExecutionResult
}

func newIdempotentService(cfg *ExecutorConfig) *idempotentService {
	return &idempotentService{
		cfg:                 cfg,
		executionResults:    make(map[string]*trax.IdempotentServiceExecutionResult),
		compensationResults: make(map[string]*trax.IdempotentServiceExecutionResult),
	}
}

func (s *idempotentService) GetIdempotentKeyExecutionStatus(
	ctx context.Context,
	sagaIdempotencyKey string,
) (trax.SagaIdempotencyKeyStatusEnum, error) {
	if s.cfg.IdempotencyBackend == "" {
		// No idempotency tracking
		return trax.SagaIdempotencyKeyStatusEnum_NotSeen, nil
	}

	_, exists := s.executionResults[sagaIdempotencyKey]
	if !exists {
		return trax.SagaIdempotencyKeyStatusEnum_NotSeen, nil
	}
	return trax.SagaIdempotencyKeyStatusEnum_Completed, nil
}

func (s *idempotentService) ExecuteSync(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) (*trax.IdempotentServiceExecutionResult, error) {
	// Check if already executed (when idempotency is enabled)
	if s.cfg.IdempotencyBackend != "" {
		result, exists := s.executionResults[sagaIdempotencyKey]
		if exists {
			return result, nil
		}
	}

	var result *trax.IdempotentServiceExecutionResult

	// Determine execution mode
	if s.cfg.ExecSimStatus != "" {
		// Simulation mode
		result = s.executeSimulation(ctx, sagaIdempotencyKey, input)
	} else {
		// Shell execution mode
		var err error
		result, err = s.executeShell(ctx, sagaIdempotencyKey, input)
		if err != nil {
			return nil, err
		}
	}

	// Store result if idempotency is enabled
	if s.cfg.IdempotencyBackend != "" {
		s.executionResults[sagaIdempotencyKey] = result
	}

	return result, nil
}

func (s *idempotentService) executeSimulation(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) *trax.IdempotentServiceExecutionResult {
	// Parse delay duration
	delay, _ := time.ParseDuration(s.cfg.ExecSimDelay)

	common.L.Info(fmt.Sprintf(
		"Simulating execution with status=%s, delay=%s, sagaIdempotencyKey=%s, input=%v",
		s.cfg.ExecSimStatus, delay, sagaIdempotencyKey, input), common.F(ctx)...)

	// Simulate delay
	time.Sleep(delay)

	var result *trax.IdempotentServiceExecutionResult

	switch s.cfg.ExecSimStatus {
	case "ok":
		// Parse result JSON
		var resultMap map[string]string
		json.Unmarshal([]byte(s.cfg.ExecSimResult), &resultMap)
		result = &trax.IdempotentServiceExecutionResult{
			Result: resultMap,
			Error:  nil,
		}
		common.L.Info(fmt.Sprintf("Simulation completed successfully: %v", resultMap), common.F(ctx)...)

	case "error":
		// Parse error JSON
		var errorMap map[string]interface{}
		json.Unmarshal([]byte(s.cfg.ExecSimError), &errorMap)
		errorMsg := fmt.Sprintf("simulated error: %v", errorMap)
		result = &trax.IdempotentServiceExecutionResult{
			Result: nil,
			Error:  errors.New(errorMsg),
		}
		common.L.Error(fmt.Sprintf("Simulation completed with error: %s", errorMsg), common.F(ctx)...)

	case "sub-saga":
		// Spawn a child saga via the SagaContext injected by the executor framework
		sagaCtx := trax.GetSagaContext(ctx)
		if sagaCtx == nil {
			common.L.Error("sub-saga mode: no SagaContext in context (executor not configured with saga submitter)", common.F(ctx)...)
			result = &trax.IdempotentServiceExecutionResult{
				Result: nil,
				Error:  errors.New("no SagaContext available for sub-saga spawning"),
			}
		} else {
			common.L.Info(fmt.Sprintf("sub-saga mode: spawning child saga '%s'", s.cfg.SubSagaTemplateId), common.F(ctx)...)
			subResult, err := sagaCtx.SpawnSubSaga(ctx, s.cfg.SubSagaTemplateId, input, sagaIdempotencyKey)
			if err != nil {
				common.L.Error(fmt.Sprintf("sub-saga '%s' failed: %v", s.cfg.SubSagaTemplateId, err), common.F(ctx)...)
				result = &trax.IdempotentServiceExecutionResult{
					Result: nil,
					Error:  fmt.Errorf("sub-saga %s failed: %w", s.cfg.SubSagaTemplateId, err),
				}
			} else {
				common.L.Info(fmt.Sprintf("sub-saga '%s' completed: instanceId=%s, state=%s",
					s.cfg.SubSagaTemplateId, subResult.SagaInstanceId, subResult.State), common.F(ctx)...)
				resultMap := map[string]string{
					"sub_saga_instance_id": subResult.SagaInstanceId,
					"sub_saga_state":       string(subResult.State),
				}
				for k, v := range subResult.Outputs {
					resultMap[k] = v
				}
				result = &trax.IdempotentServiceExecutionResult{
					Result: resultMap,
					Error:  nil,
				}
			}
		}

	case "noreturn":
		common.L.Info("Simulation with noreturn - blocking forever", common.F(ctx)...)
		// Block forever - this simulates a step that never completes
		select {}
	}

	return result
}

func (s *idempotentService) executeShell(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) (*trax.IdempotentServiceExecutionResult, error) {
	// Pre-delay
	preDelay, _ := time.ParseDuration(s.cfg.ExecShellPreDelay)
	if preDelay > 0 {
		common.L.Info(fmt.Sprintf(
			"Executing shell pre-delay: %s (sagaIdempotencyKey=%s)",
			preDelay, sagaIdempotencyKey), common.F(ctx)...)
		time.Sleep(preDelay)
	}

	common.L.Info(fmt.Sprintf(
		"Executing shell command: %s (sagaIdempotencyKey=%s)",
		s.cfg.ExecShell, sagaIdempotencyKey), common.F(ctx)...)

	// Execute shell command
	cmd := exec.CommandContext(ctx, "sh", "-c", s.cfg.ExecShell)
	cmd.Env = os.Environ()

	// Add input as environment variables
	for key, value := range input {
		cmd.Env = append(cmd.Env, fmt.Sprintf("INPUT_%s=%s", key, value))
	}

	output, err := cmd.CombinedOutput()

	// Post-delay
	postDelay, _ := time.ParseDuration(s.cfg.ExecShellPostDelay)
	if postDelay > 0 {
		common.L.Info(fmt.Sprintf(
			"Executing shell post-delay: %s (sagaIdempotencyKey=%s)",
			postDelay, sagaIdempotencyKey), common.F(ctx)...)
		time.Sleep(postDelay)
	}

	var result *trax.IdempotentServiceExecutionResult

	if err != nil {
		errorMsg := fmt.Sprintf("shell command failed: %v, output: %s", err, string(output))
		common.L.Error(errorMsg, common.F(ctx)...)
		result = &trax.IdempotentServiceExecutionResult{
			Result: nil,
			Error:  errors.New(errorMsg),
		}
	} else {
		common.L.Info(fmt.Sprintf("Shell command completed successfully, output: %s",
			string(output)), common.F(ctx)...)

		// Try to parse output as JSON, otherwise store as plain text
		resultMap := make(map[string]string)
		var jsonOutput map[string]interface{}
		if err := json.Unmarshal(output, &jsonOutput); err == nil {
			// Successfully parsed as JSON
			for k, v := range jsonOutput {
				resultMap[k] = fmt.Sprintf("%v", v)
			}
		} else {
			// Store as plain output
			resultMap["output"] = string(output)
		}

		result = &trax.IdempotentServiceExecutionResult{
			Result: resultMap,
			Error:  nil,
		}
	}

	return result, nil
}

func (s *idempotentService) ExecuteAsync(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
	cb func(result *trax.IdempotentServiceExecutionResult, err error),
) {
	go func() {
		result, err := s.ExecuteSync(ctx, sagaIdempotencyKey, input)
		cb(result, err)
	}()
}

func (s *idempotentService) GetIdempotentKeyCompensationStatus(
	ctx context.Context,
	sagaIdempotencyKey string,
) (trax.SagaIdempotencyKeyStatusEnum, error) {
	if s.cfg.IdempotencyBackend == "" {
		return trax.SagaIdempotencyKeyStatusEnum_NotSeen, nil
	}

	_, exists := s.compensationResults[sagaIdempotencyKey]
	if !exists {
		return trax.SagaIdempotencyKeyStatusEnum_NotSeen, nil
	}
	return trax.SagaIdempotencyKeyStatusEnum_Completed, nil
}

func (s *idempotentService) CompensateSync(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) (*trax.IdempotentServiceExecutionResult, error) {
	// Check if already compensated (when idempotency is enabled)
	if s.cfg.IdempotencyBackend != "" {
		result, exists := s.compensationResults[sagaIdempotencyKey]
		if exists {
			return result, nil
		}
	}

	var result *trax.IdempotentServiceExecutionResult

	// Determine compensation mode
	if s.cfg.CompSimStatus != "" {
		// Compensation simulation mode
		result = s.compensateSimulation(ctx, sagaIdempotencyKey, input)
	} else {
		// Compensation shell mode
		var err error
		result, err = s.compensateShell(ctx, sagaIdempotencyKey, input)
		if err != nil {
			return nil, err
		}
	}

	// Store result if idempotency is enabled
	if s.cfg.IdempotencyBackend != "" {
		s.compensationResults[sagaIdempotencyKey] = result
	}

	return result, nil
}

func (s *idempotentService) compensateSimulation(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) *trax.IdempotentServiceExecutionResult {
	// Parse delay duration
	delay, _ := time.ParseDuration(s.cfg.CompSimDelay)

	common.L.Info(fmt.Sprintf(
		"Simulating compensation with status=%s, delay=%s, sagaIdempotencyKey=%s, input=%v",
		s.cfg.CompSimStatus, delay, sagaIdempotencyKey, input), common.F(ctx)...)

	// Simulate delay
	time.Sleep(delay)

	var result *trax.IdempotentServiceExecutionResult

	switch s.cfg.CompSimStatus {
	case "ok":
		// Parse result JSON if provided
		var resultMap map[string]string
		if s.cfg.CompSimResult != "" {
			json.Unmarshal([]byte(s.cfg.CompSimResult), &resultMap)
		} else {
			resultMap = map[string]string{}
		}
		result = &trax.IdempotentServiceExecutionResult{
			Result: resultMap,
			Error:  nil,
		}
		common.L.Info(fmt.Sprintf("Compensation simulation completed successfully: %v",
			resultMap), common.F(ctx)...)

	case "error":
		// Parse error JSON
		var errorMap map[string]interface{}
		json.Unmarshal([]byte(s.cfg.CompSimError), &errorMap)
		errorMsg := fmt.Sprintf("simulated compensation error: %v", errorMap)
		result = &trax.IdempotentServiceExecutionResult{
			Result: nil,
			Error:  errors.New(errorMsg),
		}
		common.L.Error(fmt.Sprintf("Compensation simulation completed with error: %s",
			errorMsg), common.F(ctx)...)

	case "noreturn":
		common.L.Info("Compensation simulation with noreturn - blocking forever", common.F(ctx)...)
		// Block forever - this simulates a compensation that never completes
		select {}
	}

	return result
}

func (s *idempotentService) compensateShell(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
) (*trax.IdempotentServiceExecutionResult, error) {
	// Pre-delay
	preDelay, _ := time.ParseDuration(s.cfg.CompShellPreDelay)
	if preDelay > 0 {
		common.L.Info(fmt.Sprintf(
			"Compensation shell pre-delay: %s (sagaIdempotencyKey=%s)",
			preDelay, sagaIdempotencyKey), common.F(ctx)...)
		time.Sleep(preDelay)
	}

	common.L.Info(fmt.Sprintf(
		"Executing compensation shell command: %s (sagaIdempotencyKey=%s)",
		s.cfg.CompShell, sagaIdempotencyKey), common.F(ctx)...)

	// Execute compensation shell command
	cmd := exec.CommandContext(ctx, "sh", "-c", s.cfg.CompShell)
	cmd.Env = os.Environ()

	// Add input as environment variables
	for key, value := range input {
		cmd.Env = append(cmd.Env, fmt.Sprintf("INPUT_%s=%s", key, value))
	}

	output, err := cmd.CombinedOutput()

	// Post-delay
	postDelay, _ := time.ParseDuration(s.cfg.CompShellPostDelay)
	if postDelay > 0 {
		common.L.Info(fmt.Sprintf(
			"Compensation shell post-delay: %s (sagaIdempotencyKey=%s)",
			postDelay, sagaIdempotencyKey), common.F(ctx)...)
		time.Sleep(postDelay)
	}

	var result *trax.IdempotentServiceExecutionResult

	if err != nil {
		errorMsg := fmt.Sprintf("compensation shell command failed: %v, output: %s", err, string(output))
		common.L.Error(errorMsg, common.F(ctx)...)
		result = &trax.IdempotentServiceExecutionResult{
			Result: nil,
			Error:  errors.New(errorMsg),
		}
	} else {
		common.L.Info(fmt.Sprintf("Compensation shell command completed successfully, output: %s",
			string(output)), common.F(ctx)...)

		// Try to parse output as JSON, otherwise store as plain text
		resultMap := make(map[string]string)
		var jsonOutput map[string]interface{}
		if err := json.Unmarshal(output, &jsonOutput); err == nil {
			// Successfully parsed as JSON
			for k, v := range jsonOutput {
				resultMap[k] = fmt.Sprintf("%v", v)
			}
		} else {
			// Store as plain output
			resultMap["output"] = string(output)
		}

		result = &trax.IdempotentServiceExecutionResult{
			Result: resultMap,
			Error:  nil,
		}
	}

	return result, nil
}

func (s *idempotentService) CompensateAsync(
	ctx context.Context,
	sagaIdempotencyKey string,
	input map[string]string,
	cb func(result *trax.IdempotentServiceExecutionResult, err error),
) {
	go func() {
		result, err := s.CompensateSync(ctx, sagaIdempotencyKey, input)
		cb(result, err)
	}()
}
