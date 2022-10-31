package main

import (
	"context"
	"crypto/tls"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/ktenzer/temporal-demo-apps/moneytransfer"
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
	referenceId := uuid.New().String()
	rand.Seed(time.Now().UnixNano())
	minTransfer := float32(rand.Intn(100))
	maxTransfer := float32(rand.Intn(1000000))

	options := client.StartWorkflowOptions{
		ID:        "transfer-money-workflow-" + referenceId,
		TaskQueue: moneytransfer.TransferMoneyTaskQueue,
	}
	transferDetails := moneytransfer.TransferDetails{
		Amount:      minTransfer + maxTransfer,
		FromAccount: strconv.Itoa(rand.Intn(9999 - 1000)),
		ToAccount:   strconv.Itoa(rand.Intn(9999 - 1000)),
		ReferenceID: referenceId,
	}
	we, err := c.ExecuteWorkflow(context.Background(), options, moneytransfer.TransferMoney, transferDetails)
	if err != nil {
		log.Fatalln("error starting TransferMoney workflow", err)
	}
	printResults(transferDetails, we.GetID(), we.GetRunID())
}

func printResults(transferDetails moneytransfer.TransferDetails, workflowID, runID string) {
	log.Printf(
		"\nTransfer of $%f from account %s to account %s is processing. ReferenceID: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)
	log.Printf(
		"\nWorkflowID: %s RunID: %s\n",
		workflowID,
		runID,
	)
}
