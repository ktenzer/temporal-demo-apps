package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/ktenzer/temporal-demo-apps/moneytransfer"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	var c client.Client
	var err error
	var cert tls.Certificate

	if os.Getenv("MTLS") == "false" {
		c, err = client.Dial(client.Options{
			HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
			Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
		})
	} else {
		cert, err = tls.LoadX509KeyPair(os.Getenv("TEMPORAL_TLS_CERT"), os.Getenv("TEMPORAL_TLS_KEY"))
		if err != nil {
			log.Fatalln("Unable to load certs", err)
		}

		c, err = client.Dial(client.Options{
			HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
			Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
			ConnectionOptions: client.ConnectionOptions{
				TLS: &tls.Config{Certificates: []tls.Certificate{cert}},
			},
		})
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
