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

	// Setup logger
	logger := workflow.GetLogger(ctx)
	logger.Info("Backup workflow started", "appName", signal.AppName, "backupId", signal.BackupId)

	// Default retry policy
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:        time.Second,
		BackoffCoefficient:     2.0,
		MaximumInterval:        time.Second * 100,   // 100 * InitialInterval
		MaximumAttempts:        0,                   // Unlimited
		NonRetryableErrorTypes: []string{"bad-bug"}, // empty
	}

	// Update search attributes BackupType
	attributes := map[string]interface{}{
		"BackupType": signal.AppName,
	}
	err := workflow.UpsertSearchAttributes(ctx, attributes)
	if err != nil {
		logger.Error("Upsert search attributes failed: " + err.Error())
		return workflowResult, err
	}

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger.Info("Backup workflow started", "appName", signal.AppName, "backupId", signal.BackupId)

	// Setup query handler
	queryResult := "started"
	err = workflow.SetQueryHandler(ctx, "state", func(input []byte) (string, error) {
		return queryResult, nil
	})
	if err != nil {
		logger.Error("SetQueryHandler failed: " + err.Error())
		return workflowResult, err
	}
	queryResult = "waiting for activities to start"

	// Quiesce Activity
	workflowResult, msg, err := RunQuiesce(ctx, workflowResult)
	workflowMessages = append(workflowMessages, msg)
	workflowResult.Messages = workflowMessages
	if err != nil {
		queryResult = "failed"
		err = UpdateBackupStatusSearchAttribute(ctx, queryResult)
		if err != nil {
			return workflowResult, err
		}

		return workflowResult, err
	}
	queryResult = "quiesce suceeded"

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

		queryResult = "failed"
		err = UpdateBackupStatusSearchAttribute(ctx, queryResult)
		if err != nil {
			return workflowResult, err
		}

		return workflowResult, err
	} else if err != nil {
		queryResult = "failed"
		err = UpdateBackupStatusSearchAttribute(ctx, queryResult)
		if err != nil {
			return workflowResult, err
		}

		return workflowResult, err
	}
	queryResult = "backup succeeded"

	// UnQuiesce Activity
	workflowResult, msg, err = RunUnQuiesce(ctx, workflowResult)
	workflowMessages = append(workflowMessages, msg)
	workflowResult.Messages = workflowMessages

	if err != nil {
		queryResult = "failed"
		err = UpdateBackupStatusSearchAttribute(ctx, queryResult)
		if err != nil {
			return workflowResult, err
		}

		return workflowResult, err
	}
	queryResult = "succeeded"
	err = UpdateBackupStatusSearchAttribute(ctx, queryResult)
	if err != nil {
		return workflowResult, err
	}

	logger.Info("Backup workflow completed.", "result", workflowResult)
	return workflowResult, nil
}

func UpdateBackupStatusSearchAttribute(ctx workflow.Context, queryResult string) error {
	attributes := map[string]interface{}{
		"BackupStatus": queryResult,
	}
	workflow.UpsertSearchAttributes(ctx, attributes)
	err := workflow.UpsertSearchAttributes(ctx, attributes)
	if err != nil {
		return err
	}

	return nil
}
