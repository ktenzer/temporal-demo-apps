package main

import (
	"log"

	"github.com/ktenzer/temporal-demo-apps/moneytransfer"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, moneytransfer.TransferMoneyTaskQueue, worker.Options{})
	w.RegisterWorkflow(moneytransfer.TransferMoney)
	w.RegisterActivity(moneytransfer.Withdraw)
	w.RegisterActivity(moneytransfer.WithdrawCompensation)
	w.RegisterActivity(moneytransfer.Deposit)
	w.RegisterActivity(moneytransfer.DepositCompensation)
	w.RegisterActivity(moneytransfer.StepWithError)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
