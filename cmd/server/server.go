package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sebnyberg/autodotenv"
	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/policefeed/feed"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type serverConfig struct {
	Addr    string `value:"localhost:0" usage:"Address, random port is allocated when zero"`
	Regions string `value:"" usage:"comma-separated list of region IDs from the Swedish Police Website"`
	feed.DBConfig
}

func NewServerCmd() *cli.Command {
	var conf serverConfig

	if _, err := autodotenv.LoadDotenvIfExists(); err != nil {
		log.Fatalln(err)
	}

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
	// Database connection setup & check
	db, err := conf.DBConfig.OpenDB()
	if err != nil {
		return fmt.Errorf("open database conn err, %w", err)
	}
	if err := feed.ValidateSchema(db); err != nil {
		return fmt.Errorf("validate database schema err, %w", err)
	}

	// Create RSS feed fetcher
	rssFeed := &feed.RSSAdapter{
		RegionIDs: strings.Split(conf.Regions, ","),
	}

	// Init database eventStorage
	eventStorage := feed.NewEventStorage(db)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return feed.Update(ctx, rssFeed, eventStorage, time.Second*10)
	})

	return g.Wait()
}
