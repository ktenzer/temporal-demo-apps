package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"

	"github.com/google/uuid"
	"github.com/temporal-demo-apps/backup"
	"go.temporal.io/sdk/client"
)

type WorkflowResult struct {
	Code     int      `json:"code"`
	Messages []string `json:"message"`
}

func main() {
	// The client is a heavyweight object that should be created once per process.
	const clientCertPath string = "/home/ktenzer/temporal/certs/ca.pem"
	const clientKeyPath string = "/home/ktenzer/temporal/certs/ca.key"

	var c client.Client
	var err error
	var cert tls.Certificate

	_, isMTLS := os.LookupEnv("MTLS")
	if isMTLS {

		cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
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
	} else {
		c, err = client.Dial(client.Options{})
	}

	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	backupId := uuid.New().String()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "backup_sample_" + backupId,
		TaskQueue: "backup-sample",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, backup.Workflow, backupId)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "BackupId", backupId, "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result WorkflowResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
