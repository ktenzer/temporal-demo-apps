package backup

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

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
