package versioning

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Verisoning workflow started", "name", name)

	// Setup query handler
	workflowState := "running"
	err := workflow.SetQueryHandler(ctx, "state", func(input []byte) (string, error) {
		return workflowState, nil
	})
	if err != nil {
		logger.Error("SetQueryHandler failed: " + err.Error())
		return "", err
	}

	var result string
	v := workflow.GetVersion(ctx, "Version", workflow.DefaultVersion, 2)
	if v == workflow.DefaultVersion {
		err := workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}
	} else if v == 1 {
		err := workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}
	} else {
		err := workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}

		err = workflow.ExecuteActivity(ctx, ActivityC).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return "", err
		}

		// Example for showing replaying and non-determinism
		// add new activity
		//err = workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, &result)
		// if err != nil {
		//	logger.Error("Activity failed.", "Error", err)
		//	return "", err
		//}

		// set timer for 5 minutes
		workflow.Sleep(ctx, time.Minute*1)
	}

	workflowState = "completed"
	logger.Info("Versioning workflow completed.", "result", result)

	return result, nil
}

func ActivityA(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ActivityA")
	return "Running Activity A", nil
}

func ActivityB(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ActivityB")
	return "Running Activity B", nil
}

func ActivityC(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ActivityC")
	return "Running Activity C", nil
}
