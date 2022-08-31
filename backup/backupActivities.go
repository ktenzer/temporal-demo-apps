package backup

import (
	"context"
	"errors"
	"strconv"

	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func QuiesceActivity(ctx context.Context, backupId string) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Activity", "name", backupId)

	result, err := Quiesce("localhost", "9977", backupId)
	if err != nil {
		return "HTTP error", err
	}

	if result.Code != 0 {
		return "Result message " + result.Message + "!", errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return "Result message " + result.Message + "!", nil
}

func BackupActivity(ctx context.Context, backupId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := Backup("localhost", "9977", backupId)
	if err != nil {
		return "HTTP error", err
	}

	if result.Code != 0 {
		return "Result message " + result.Message + "!", errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return "Result message " + result.Message + "!", nil
}

func UnQuiesceActivity(ctx context.Context, backupId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", backupId)

	result, err := UnQuiesce("localhost", "9977", backupId)
	if err != nil {
		return "HTTP error", err
	}

	if result.Code != 0 {
		return "Result message " + result.Message + "!", errors.New("Error Code: " + strconv.Itoa(result.Code))
	}

	return "Result message " + result.Message + "!", nil
}
