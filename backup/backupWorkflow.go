package backup

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func Workflow(ctx workflow.Context, backupId string) (string, error) {
	var result string

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
	backupState, _ := GetBackupState("localhost", "9977", backupId)
	if backupState == "" || backupState == "unquiesced" {
		customQuiesceAO := SetCustomRetryPolicy(1, 10, 10)
		customQuiesceCTX := workflow.WithActivityOptions(ctx, customQuiesceAO)
		logger = workflow.GetLogger(customQuiesceCTX)

		err := workflow.ExecuteActivity(customQuiesceCTX, QuiesceActivity, backupId).Get(customQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}
	}

	// Backup Activity
	// want to run unquiesce if max retries or duration is reached?
	backupState, _ = GetBackupState("localhost", "9977", backupId)
	if backupState == "" || backupState == "quiesced" {
		customBackupAO := SetCustomRetryPolicy(1, 10, 1)
		customBackupCTX := workflow.WithActivityOptions(ctx, customBackupAO)
		logger = workflow.GetLogger(customBackupCTX)

		err := workflow.ExecuteActivity(customBackupCTX, BackupActivity, backupId).Get(customBackupCTX, &result)

		var timeoutErr *temporal.ApplicationError
		if errors.As(err, &timeoutErr) {
			// UnQuiesce Activity
			backupState, _ = GetBackupState("localhost", "9977", backupId)
			if backupState == "backup" || backupState == "quiesced" {
				customUnQuiesceAO := SetCustomRetryPolicy(1, 10, 10)
				customUnQuiesceCTX := workflow.WithActivityOptions(ctx, customUnQuiesceAO)
				logger = workflow.GetLogger(customUnQuiesceCTX)

				err := workflow.ExecuteActivity(customUnQuiesceCTX, UnQuiesceActivity, backupId).Get(customUnQuiesceCTX, &result)
				if err != nil {
					logger.Error("Activity failed.", "Error", err)
					return "", err
				}
			}
		}

		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}
	}

	// UnQuiesce Activity
	backupState, _ = GetBackupState("localhost", "9977", backupId)
	if backupState == "backup" || backupState == "quiesced" {
		customUnQuiesceAO := SetCustomRetryPolicy(1, 10, 10)
		customUnQuiesceCTX := workflow.WithActivityOptions(ctx, customUnQuiesceAO)
		logger = workflow.GetLogger(customUnQuiesceCTX)

		err := workflow.ExecuteActivity(customUnQuiesceCTX, UnQuiesceActivity, backupId).Get(customUnQuiesceCTX, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}
	}

	logger.Info("Backup workflow completed.", "result", result)

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
