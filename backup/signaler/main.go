package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type BackupSignal struct {
	Action   string `json:"action"`
	AppName  string `json:"appName"`
	BackupId string `json:"backupId"`
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

	// loop through a bunch of apps, creating signal for each
	apps := []string{"Oracle", "MySQL", "PostgreSQL", "Cassandra", "Couchbase"}
	for _, app := range apps {
		backupId := uuid.New().String()

		signal := BackupSignal{
			Action:   "RunBackup",
			AppName:  app,
			BackupId: backupId,
		}
		err = c.SignalWorkflow(context.Background(), "backup_sample", "", "start-backup", signal)
		if err != nil {
			log.Fatalln("Error sending the Signal", err)
			return
		}

		log.Println("Started workflow", "AppName", app, "BackupId", backupId)
	}
}
