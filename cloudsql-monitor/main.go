package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cli "github.com/urfave/cli/v2"

	"github.com/themdoan/devops-utilities/cloudtrace"
)

func main() {
	if err := run(os.Args); err != nil {
		panic(err)
	}
}

func run(args []string) error {
	// var err error
	// cfg, err = config.Load()
	// if err != nil {
	// 	return err
	// }
	if len(args) > 0 {
		fmt.Printf("number of arg: %d \n", len(args))
	}
	app := cli.NewApp()
	app.Name = "service"
	app.Commands = []*cli.Command{
		{
			Name:   "hello",
			Usage:  "hello world",
			Action: serverAction,
		},
		{
			Name:   "cloudsql",
			Usage:  "cloudsql alert",
			Action: cloudsqlMonitor,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "project-id",
					Value:   "gcloud project-id",
					Aliases: []string{"p"},
					Usage:   "Specify google cloud project-id",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
	return nil
}

func serverAction(cliCtx *cli.Context) error {
	fmt.Printf("boom! I say!: %s", cliCtx.String("project"))
	return nil
}
func cloudsqlMonitor(cliCtx *cli.Context) error {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "", log.Lshortfile)
	)
	var client_err error
	start := time.Now()

	projectID := cliCtx.String("project-id")

	const intervalTime = time.Hour * 6 * 1

	listCtx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Minute*15))
	defer func() {
		cancel()
		logger.Print("Finished testConnection", "duration", time.Since(start).String())
	}()
	client, client_err := cloudtrace.NewClientWithGCE(listCtx)
	if client_err != nil {
		logger.Printf("failed to init client: %s", client_err)
	}

	entries, client_err := client.ListTraces(listCtx, &cloudtrace.TracesQuery{
		ProjectID: projectID,
		Limit:     10,
		Filter:    "latency:3s",
		TimeRange: cloudtrace.TimeRange{
			From: time.Now().Add(-intervalTime),
			To:   time.Now(),
		},
	})
	if client_err != nil {
		logger.Printf("failed to listTraces: %s", client_err)
	}

	// logger.Printf("ListTraces: %v", entries)
	if len(entries) == 0 {
		fmt.Printf("Slow query: 0")
		return nil
	}

	alertmanager_url := os.Getenv("ALERTMANAGER_URL")

	if alertmanager_url == "" {
		panic("missing env ALERTMANAGER_URL")
	}

	alertmanager, _ := cloudtrace.NewAlertmanager(alertmanager_url, "", nil)

	err := alertmanager.Post(context.TODO(), entries)
	if err != nil {
		logger.Printf("alert post fail: %s", err)
	}
	fmt.Print(&buf)
	return nil
}
