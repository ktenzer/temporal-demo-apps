package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"

	"github.com/google/uuid"
	"github.com/ktenzer/temporal-demo-apps/backup"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.

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

	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	app := "Oracle"
	backupId := uuid.New().String()

	err = SendSignal(c, app, backupId)
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}
}

func SendSignal(c client.Client, app, backupId string) error {
	workflowId := "backup_sample_" + app + "_" + backupId
	signal := backup.BackupSignal{
		Action:   "RunBackup",
		AppName:  app,
		BackupId: backupId,
	}

	err := c.SignalWorkflow(context.Background(), "backup_sample", "", "start-backup", signal)
	if err != nil {
		return err
	}

	log.Println("Workflow[" + workflowId + "]")

	return nil
}
func SendQuery(c client.Client, workflowId string) interface{} {

	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", "state")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	return result
}
