package server

import (
	"context"
	"net"
	"os"
	"os/signal"

	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type serverConfig struct {
	Addr    string `value:"localhost:0" usage:"Address, random port is allocated when zero"`
	Regions string `value:"stockholms-lan" usage:"comma-separated list of region IDs from the Swedish Police Website"`
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
	return nil
	// // Find valid address
	// listener, err := net.Listen("tcp", conf.Addr)
	// if err != nil {
	// 	return fmt.Errorf("start server err, %v", err)
	// }
	// // addr := listener.Addr()

	// // Validate regions provided in config
	// regions := feed.NewRegions(conf.RSSBaseURL)
	// for _, region := range strings.Split(conf.Regions, ",") {
	// 	if !regions.Exists(region) {
	// 		return fmt.Errorf(
	// 			"unknown region %v, choose one or more of %v",
	// 			strings.Join(regions.ListIDs(), ","),
	// 		)
	// 	}
	// }
	// _ = listener
	// return nil

	//
}
