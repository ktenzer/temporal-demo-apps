package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"

	"github.com/temporal-demo-apps/backup"
	"go.temporal.io/sdk/client"
)

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

	app := "MySQL"
	backupId := "81e68a74-c2ac-438f-a8ce-3d6cba60c754"

	err = SendSignal(c, app, backupId)
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}
}

func SendSignal(c client.Client, app, backupId string) error {
	workflowId := "backup_sample_" + app + "_" + backupId
	signal := backup.BackupSignal{
		Action:       "ScheduleBackup",
		AppName:      app,
		BackupId:     backupId,
		CronSchedule: "* * * * *",
	}

	err := c.SignalWorkflow(context.Background(), "backup_sample", "", "start-backup", signal)
	if err != nil {
		return err
	}

	log.Println("Workflow[" + workflowId + "] Scheduled [" + signal.CronSchedule + "]")

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
