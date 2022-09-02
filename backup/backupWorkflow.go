package backup

import (
	"errors"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func Workflow(ctx workflow.Context, backupId string) (string, error) {

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, backupId)
	// Wait for the Child Workflow Execution to spawn
	var childWE workflow.Execution
	if err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		return "Childworkflow failed to start!", err
	}

	return "Childworkflow started successfully", nil
}

func ChildWorkflow(ctx workflow.Context, backupId string) (WorkflowResult, error) {
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
