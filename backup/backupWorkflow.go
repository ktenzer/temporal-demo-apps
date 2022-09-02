package backup

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func Workflow(ctx workflow.Context, backupId string) (WorkflowResult, error) {
	var workflowResult WorkflowResult
	var workflowMessages []string
	workflowResult.Id = backupId

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
	workflowResult, msg, err := RunQuiesce(ctx, workflowResult)
	workflowMessages = append(workflowMessages, msg)
	workflowResult.Messages = workflowMessages
	if err != nil {
		return workflowResult, err
	}

	// Backup Activity
	// unquiesce if we return timeout or app error
	workflowResult, msg, err = RunBackup(ctx, workflowResult)
	workflowMessages = append(workflowMessages, msg)
	workflowResult.Messages = workflowMessages

	var timeoutErr *temporal.TimeoutError
	var appErr *temporal.ApplicationError

	if errors.As(err, &appErr) || errors.As(err, &timeoutErr) {
		// UnQuiesce Activity
		workflowResult, msg, err := RunUnQuiesce(ctx, workflowResult)
		workflowMessages = append(workflowMessages, msg)
		workflowResult.Messages = workflowMessages

		return workflowResult, err
	} else if err != nil {
		return workflowResult, err
	}

	// UnQuiesce Activity
	workflowResult, msg, err = RunUnQuiesce(ctx, workflowResult)
	workflowMessages = append(workflowMessages, msg)
	workflowResult.Messages = workflowMessages

	if err != nil {
		return workflowResult, err
	}

	logger.Info("Backup workflow completed.", "result", workflowResult)
	return workflowResult, nil
}

func RunQuiesce(ctx workflow.Context, workflowResult WorkflowResult) (WorkflowResult, string, error) {
	var result Result

	if workflowResult.State == "" || workflowResult.State == "unquiesced" {
		customQuiesceAO := SetCustomRetryPolicy(1, 30, 30)
		customQuiesceCTX := workflow.WithActivityOptions(ctx, customQuiesceAO)
		logger := workflow.GetLogger(customQuiesceCTX)

		err := workflow.ExecuteActivity(customQuiesceCTX, QuiesceActivity, workflowResult.Id).Get(customQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return workflowResult, result.Message, err
		}

		workflowResult.Code = result.Code
		workflowResult.State = "quiesced"
	}

	return workflowResult, result.Message, nil
}

func RunBackup(ctx workflow.Context, workflowResult WorkflowResult) (WorkflowResult, string, error) {
	var result Result

	if workflowResult.State == "" || workflowResult.State == "quiesced" {
		customBackupAO := SetCustomRetryPolicy(1, 10, 1)
		customUnBackupCTX := workflow.WithActivityOptions(ctx, customBackupAO)
		logger := workflow.GetLogger(customUnBackupCTX)

		err := workflow.ExecuteActivity(customUnBackupCTX, BackupActivity, workflowResult.Id).Get(customUnBackupCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return workflowResult, result.Message, err
		}

		workflowResult.Code = result.Code
		workflowResult.State = "backup"
	}

	return workflowResult, result.Message, nil
}

func RunUnQuiesce(ctx workflow.Context, workflowResult WorkflowResult) (WorkflowResult, string, error) {
	var result Result

	if workflowResult.State == "backup" || workflowResult.State == "quiesced" {
		customUnQuiesceAO := SetCustomRetryPolicy(1, 30, 30)
		customUnQuiesceCTX := workflow.WithActivityOptions(ctx, customUnQuiesceAO)
		logger := workflow.GetLogger(customUnQuiesceCTX)

		err := workflow.ExecuteActivity(customUnQuiesceCTX, UnQuiesceActivity, workflowResult.Id).Get(customUnQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return workflowResult, result.Message, err
		}

		workflowResult.Code = result.Code
		workflowResult.State = "unquiesced"
	}

	return workflowResult, result.Message, nil
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
