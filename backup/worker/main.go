package main

import (
	"log"

	"crypto/tls"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/ktenzer/temporal-demo-apps/backup"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.

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

	w := worker.New(c, "backup-sample", worker.Options{})

	w.RegisterWorkflow(backup.ChildWorkflow)
	w.RegisterWorkflow(backup.SignalWorkflow)
	w.RegisterActivity(backup.QuiesceActivity)
	w.RegisterActivity(backup.BackupActivity)
	w.RegisterActivity(backup.UnQuiesceActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
