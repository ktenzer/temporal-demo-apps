package backup

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func BackupWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, QuiesceActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, BackupActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, UnQuiesceActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("Backup workflow completed.", "result", result)

	return result, nil
}

func QuiesceActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	result, err := Quiesce("localhost", "9977")
	if err != nil {
		return "Result message " + result.Message + "!", err
	} else {
		return "Result message " + result.Message + "!", nil
	}
}

func BackupActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	result, err := Backup("localhost", "9977")
	if err != nil {
		return "Result message " + result.Message + "!", err
	} else {
		return "Result message " + result.Message + "!", nil
	}
}

func UnQuiesceActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	result, err := UnQuiesce("localhost", "9977")
	if err != nil {
		return "Result message " + result.Message + "!", err
	} else {
		return "Result message " + result.Message + "!", nil
	}
}
