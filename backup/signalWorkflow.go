package backup

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func SignalWorkflow(ctx workflow.Context) (string, error) {
	var workflowResult WorkflowResult
	//var workflowMessages []string
	workflowResult.Id = ""

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

	//logger := workflow.GetLogger(ctx)

	var signal BackupSignal
	signalChan := workflow.GetSignalChannel(ctx, "start-backup")
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
	})
	selector.Select(ctx)
	if len(signal.Action) == 0 || len(signal.AppName) == 0 || len(signal.BackupId) == 0 {
		return "error", errors.New("Received empty signal")
	}

	if signal.Action == "RunBackup" {
		err := SpawnManualChildWorkflow(ctx, signal)
		if err != nil {
			return "Childworkflow failed to start!", err
		}

		err = DrainManualSignals(ctx, signal, signalChan)
		if err != nil {
			return "Childworkflow failed to start!", err
		}
	} else if signal.Action == "ScheduleBackup" {
		err := SpawnScheduleChildWorkflow(ctx, signal)
		if err != nil {
			return "Childworkflow failed to start!", err
		}

		err = DrainScheduledSignals(ctx, signal, signalChan)
		if err != nil {
			return "Childworkflow failed to start!", err
		}
	} else {
		return "error", errors.New("Received incorrect signal, only RunBackup and ScheduleBackup are supported!")
	}

	return "success", workflow.NewContinueAsNewError(ctx, SignalWorkflow)
}

func SpawnManualChildWorkflow(ctx workflow.Context, signal BackupSignal) error {
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:        "backup_sample_" + signal.AppName + "_" + signal.BackupId,
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
	}

	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, signal)
	// Wait for the Child Workflow Execution to spawn
	var childWE workflow.Execution
	if err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		return err
	}

	return nil
}

func SpawnScheduleChildWorkflow(ctx workflow.Context, signal BackupSignal) error {
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:        "backup_sample_" + signal.AppName + "_" + signal.BackupId,
		CronSchedule:      signal.CronSchedule,
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
	}

	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, signal)
	// Wait for the Child Workflow Execution to spawn
	var childWE workflow.Execution

	//TODO need check if workflow is already running
	if err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		fmt.Println("ChildWorkflow execution error[" + err.Error() + "]")
	}

	return nil
}

func DrainManualSignals(ctx workflow.Context, signal BackupSignal, signalChan workflow.ReceiveChannel) error {
	// drain any signals that come during continue-as-new
	for {
		ok := signalChan.ReceiveAsync(&signal)

		if !ok {
			break
		}
		workflow.GetLogger(ctx).Info("Received signal!", "signal", "start-backup", "RunBackup", signal)

		// execute child worklflow
		err := SpawnManualChildWorkflow(ctx, signal)
		if err != nil {
			return err
		}
	}

	return nil
}

func DrainScheduledSignals(ctx workflow.Context, signal BackupSignal, signalChan workflow.ReceiveChannel) error {
	// drain any signals that come during continue-as-new
	for {
		ok := signalChan.ReceiveAsync(&signal)

		if !ok {
			break
		}
		workflow.GetLogger(ctx).Info("Received signal!", "signal", "start-backup", "ScheduleBackup", signal)

		// execute child worklflow
		err := SpawnScheduleChildWorkflow(ctx, signal)
		if err != nil {
			return err
		}
	}

	return nil
}
