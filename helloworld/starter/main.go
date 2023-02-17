package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"
	"crypto/x509"

	"github.com/ktenzer/temporal-demo-apps/helloworld"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.

	clientOptions := client.Options{
		HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
		Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
	}

	if os.Getenv("TEMPORAL_MTLS_TLS_CERT") != "" && os.Getenv("TEMPORAL_MTLS_TLS_KEY") != "" {
		if os.Getenv("TEMPORAL_MTLS_TLS_CA") != "" {
			caCert, err := os.ReadFile(os.Getenv("TEMPORAL_MTLS_TLS_CA"))
			if err != nil {
				log.Fatalln("failed reading server CA's certificate", err)
			}

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caCert) {
				log.Fatalln("failed to add server CA's certificate", err)
			}

			cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
			if err != nil {
				log.Fatalln("Unable to load certs", err)
			}

			var serverName string
			if os.Getenv("TEMPORAL_MTLS_TLS_ENABLE_HOST_VERIFICATION") == "true" {
				serverName = os.Getenv("TEMPORAL_MTLS_TLS_SERVER_NAME")
			}

			clientOptions.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{
					RootCAs:      certPool,
					Certificates: []tls.Certificate{cert},
					ServerName:   serverName,
				},
			}
		} else {
			cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
			if err != nil {
				log.Fatalln("Unable to load certs", err)
			}

			clientOptions.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			}
		}
	}

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
