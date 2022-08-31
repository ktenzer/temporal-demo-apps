package backup

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

const (
	BackupSignalName = "startbackup"
)

func Workflow(ctx workflow.Context, backupId string) (WorkflowResult, error) {
	var workflowResult WorkflowResult
	var workflowMessages []string
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:        time.Second,
		BackoffCoefficient:     2.0,
		MaximumInterval:        time.Second * 100,   // 100 * InitialInterval
		MaximumAttempts:        0,                   // Unlimited
		NonRetryableErrorTypes: []string{"bad-bug"}, // empty
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "backupId", backupId)

	// Quiesce Activity
	result, err := RunQuiesce(ctx, backupId)
	workflowResult.Code = result.Code
	workflowMessages = append(workflowMessages, result.Message)
	workflowResult.Messages = workflowMessages
	if err != nil {
		return workflowResult, err
	}

	// Backup Activity
	// unquiesce if we return timeout or app error
	result, err = RunBackup(ctx, backupId)
	workflowResult.Code = result.Code
	workflowMessages = append(workflowMessages, result.Message)
	workflowResult.Messages = workflowMessages

	var timeoutErr *temporal.TimeoutError
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) || errors.As(err, &timeoutErr) {
		// UnQuiesce Activity
		result, err := RunUnQuiesce(ctx, backupId)
		workflowMessages = append(workflowMessages, result.Message)
		workflowResult.Messages = workflowMessages
		return workflowResult, err
	} else if err != nil {
		return workflowResult, err
	}

	// UnQuiesce Activity
	result, err = RunUnQuiesce(ctx, backupId)
	workflowMessages = append(workflowMessages, result.Message)
	workflowResult.Messages = workflowMessages
	if err != nil {
		return workflowResult, err
	}

	logger.Info("Backup workflow completed.", "result", result)

	return workflowResult, nil
}

func RunQuiesce(ctx workflow.Context, backupId string) (Result, error) {
	var result Result
	backupState, _ := GetBackupState("localhost", "9977", backupId)
	if backupState == "" || backupState == "unquiesced" {
		customQuiesceAO := SetCustomRetryPolicy(1, 10, 10)
		customQuiesceCTX := workflow.WithActivityOptions(ctx, customQuiesceAO)
		logger := workflow.GetLogger(customQuiesceCTX)

		err := workflow.ExecuteActivity(customQuiesceCTX, QuiesceActivity, backupId).Get(customQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return result, err
		}
	}

	return result, nil
}

func RunBackup(ctx workflow.Context, backupId string) (Result, error) {
	var result Result
	backupState, _ := GetBackupState("localhost", "9977", backupId)
	if backupState == "" || backupState == "quiesced" {
		customBackupAO := SetCustomRetryPolicy(1, 10, 1)
		customUnBackupCTX := workflow.WithActivityOptions(ctx, customBackupAO)
		logger := workflow.GetLogger(customUnBackupCTX)

		err := workflow.ExecuteActivity(customUnBackupCTX, BackupActivity, backupId).Get(customUnBackupCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return result, err
		}
	}

	return result, nil
}

func RunUnQuiesce(ctx workflow.Context, backupId string) (Result, error) {
	var result Result
	backupState, _ := GetBackupState("localhost", "9977", backupId)
	if backupState == "backup" || backupState == "quiesced" {
		customUnQuiesceAO := SetCustomRetryPolicy(1, 10, 10)
		customUnQuiesceCTX := workflow.WithActivityOptions(ctx, customUnQuiesceAO)
		logger := workflow.GetLogger(customUnQuiesceCTX)

		err := workflow.ExecuteActivity(customUnQuiesceCTX, UnQuiesceActivity, backupId).Get(customUnQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return result, err
		}
	}

	return result, nil
}

func SetCustomRetryPolicy(retryIntervalSeconds, maxIntervalSeconds, maxAttempts int) workflow.ActivityOptions {

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:        time.Second * time.Duration(retryIntervalSeconds),
		BackoffCoefficient:     1.0,
		MaximumInterval:        time.Second * time.Duration(maxIntervalSeconds),
		MaximumAttempts:        int32(maxAttempts),  // Unlimited
		NonRetryableErrorTypes: []string{"bad-bug"}, // empty
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}

	return ao
}
