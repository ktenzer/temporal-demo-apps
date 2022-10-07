package backup

import (
	"context"
	"strconv"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func QuiesceActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Activity", "name", backupId)

	result, err := Quiesce("localhost", "9977", backupId)
	if err != nil {
		return result, temporal.NewApplicationError("quiesce failed", "quiesce", err)
	}

	if result.Code != 0 {
		return result, temporal.NewApplicationError("Error Code: "+strconv.Itoa(result.Code), "backup")
	}

	return result, nil
}

func BackupActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := Backup("localhost", "9977", backupId)
	if err != nil {
		return result, temporal.NewApplicationError("backup failed", "backup", err.Error())
	}

	if result.Code != 0 {
		return result, temporal.NewApplicationError("Error Code: "+strconv.Itoa(result.Code), "backup")
	}

	return result, nil
}

func UnQuiesceActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := UnQuiesce("localhost", "9977", backupId)
	if err != nil {
		return result, temporal.NewApplicationError("unquiesce failed", "unquiesce", err.Error())
	}

	if result.Code != 0 {
		return result, temporal.NewApplicationError("Error Code: "+strconv.Itoa(result.Code), "backup")
	}

	return result, nil
}
