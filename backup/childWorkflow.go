package backup

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func ChildWorkflow(ctx workflow.Context, signal BackupSignal) (WorkflowResult, error) {
	var workflowResult WorkflowResult
	var workflowMessages []string
	workflowResult.Id = signal.BackupId
	workflowResult.AppName = signal.AppName

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
	logger.Info("HelloWorld workflow started", "appName", signal.AppName, "backupId", signal.BackupId)

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
