package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/sebnyberg/flagtags"
	policefeed "github.com/sebnyberg/police-feed-se"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type serverConfig struct {
	Addr       string `value:"localhost:0" usage:"Address, random port is allocated when zero"`
	RSSBaseURL string `value:"https://polisen.se/aktuellt/rss/stockholms-lan/handelser-rss---%v/" usage:"base URL with %v placeholder for region ID"`
	Regions    string `value:"stockholms-lan" usage:"comma-separated list of region IDs from the Swedish Police Website"`
}

func NewServerCmd() *cli.Command {
	var conf serverConfig

	return &cli.Command{
		Name:        "server",
		Usage:       "start the police feed server",
		Description: "Start the server on the provided address.",
		Action: func(*cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
			defer cancel()
			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				return runServer(ctx, conf)
			})
			return g.Wait()
		},
		Flags: flagtags.MustParseFlags(&conf),
	}
}

type server struct {
	addr net.Addr
}

func runServer(ctx context.Context, conf serverConfig) error {
	// Find valid address
	listener, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		return fmt.Errorf("start server err, %v", err)
	}
	// addr := listener.Addr()

	// Validate regions provided in config
	regions := policefeed.NewRegions(conf.RSSBaseURL)
	for _, region := range strings.Split(conf.Regions, ",") {
		if !regions.Exists(region) {
			return fmt.Errorf(
				"unknown region %v, choose one or more of %v",
				strings.Join(regions.ListIDs(), ","),
			)
		}
	}
	_ = listener
	return nil

	//
}
