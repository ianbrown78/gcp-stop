package cmd

import (
	"log"
	"os"

	"gcp-stop/config"
	"gcp-stop/gcp"
	"github.com/urfave/cli/v2"
)

func Command() {
	app := &cli.App{
		Usage:     "The GCP project shutdown tool",
		Version:   "v0.0.1",
		UsageText: "e.g. gcp-stop --project test-stop-123456 --dryrun",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "project, p",
				Usage:    "GCP project id to nuke (required)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dryrun, d",
				Usage: "Perform a dryrun instead",
			},
			&cli.IntFlag{
				Name:  "timeout, t",
				Value: 400,
				Usage: "Timeout for shutdown of a single resource in seconds",
			},
			&cli.IntFlag{
				Name:  "polltime, o",
				Value: 10,
				Usage: "Time for polling resource shutdown status in seconds",
			},
		},
		Action: func(c *cli.Context) error {

			// Behaviour to stop all resources in parallel in one project at a time - will be made into loop / concurrenct project stop if required
			config := config.Config{
				Project:  c.String("project"),
				DryRun:   c.Bool("dryrun"),
				Timeout:  c.Int("timeout"),
				PollTime: c.Int("polltime"),
				Context:  gcp.Ctx,
				Zones:    gcp.GetZones(gcp.Ctx, c.String("project")),
				Regions:  gcp.GetRegions(gcp.Ctx, c.String("project")),
			}
			log.Printf("[Info] Timeout %v seconds. Polltime %v seconds. Dry run: %v", config.Timeout, config.PollTime, config.DryRun)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
