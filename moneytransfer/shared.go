package moneytransfer

import (
	"errors"
	"math/rand"
	"os"
	"time"

	"go.temporal.io/sdk/temporal"
)

const TransferMoneyTaskQueue = "TRANSFER_MONEY_TASK_QUEUE"

type TransferDetails struct {
	Amount      float32
	FromAccount string
	ToAccount   string
	ReferenceID string
}

func ChaosMonkey() error {
	if os.Getenv("ENABLE_CHAOS_MONKEY") == "true" {
		sleepTimer := rand.Intn(1250)
		time.Sleep(time.Duration(sleepTimer) * time.Millisecond)

		var code int
		rand.Seed(time.Now().UnixNano())
		min := 0
		max := 2
		code = rand.Intn(max-min+1) + min

		if code != 0 {
			return errors.New("Transaction timeout, please retry")
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func SetRetryPolicy(maxAttempts int) *temporal.RetryPolicy {

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    int32(maxAttempts),
	}

	return retryPolicy
}
