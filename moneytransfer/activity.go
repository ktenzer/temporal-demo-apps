package moneytransfer

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
)

func Withdraw(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nWithdrawing $%f from account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ReferenceID,
	)

	err := ChaosMonkey()
	if err != nil {
		fmt.Printf(
			"\nWithdraw failed! ReferenceId: %s. Error: "+err.Error()+"\n",
			transferDetails.ReferenceID,
		)
	}

	return err
}

func WithdrawCompensation(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nWithdrawing compensation $%f from account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func Deposit(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nDepositing $%f into account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)

	err := ChaosMonkey()
	if err != nil {
		fmt.Printf(
			"\nDeposit failed! ReferenceId: %s. Error: "+err.Error()+"\n",
			transferDetails.ReferenceID,
		)
	}

	return err
}

func DepositCompensation(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nDepositing compensation $%f into account %s. ReferenceId: %s\n",
		transferDetails.Amount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)

	return nil
}

func StepWithError(ctx context.Context, transferDetails TransferDetails) error {
	fmt.Printf(
		"\nSimulate failure to trigger compensation. ReferenceId: %s\n",
		transferDetails.ReferenceID,
	)

	return temporal.NewNonRetryableApplicationError("Account data mismatch", "StepError", errors.New("data mismatch"))
}
