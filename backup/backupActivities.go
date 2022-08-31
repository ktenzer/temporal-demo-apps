package backup

import (
	"context"
	"errors"
	"strconv"

	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func QuiesceActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Activity", "name", backupId)

	result, err := Quiesce("localhost", "9977", backupId)
	if err != nil {
		return result, err
	}

	if result.Code != 0 {
		return result, errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return result, nil
}

func BackupActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := Backup("localhost", "9977", backupId)
	if err != nil {
		return result, err
	}

	if result.Code != 0 {
		return result, errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return result, nil
}

func UnQuiesceActivity(ctx context.Context, backupId string) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := UnQuiesce("localhost", "9977", backupId)
	if err != nil {
		return result, err
	}

	if result.Code != 0 {
		return result, errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return result, nil
}
