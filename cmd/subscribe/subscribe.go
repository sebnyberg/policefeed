package subscribe

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/policefeed/feed"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type subscriberConfig struct {
	Regions string `value:"" usage:"comma-separated list of region IDs from the Swedish Police Website. If 'all' subscribes to all regions."`
}

func NewSubscribeCmd() *cli.Command {
	var conf subscriberConfig

	return &cli.Command{
		Name:        "subscribe",
		Usage:       "subscribe to the police feed",
		Description: "Subscribe to events occurring on the police feed",
		Action: func(*cli.Context) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
			defer cancel()
			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				return runSubscribe(ctx, conf)
			})
			return g.Wait()
		},
		Flags: flagtags.MustParseFlags(&conf),
	}
}

func runSubscribe(ctx context.Context, conf subscriberConfig) error {
	regionIDs := strings.Split(conf.Regions, ",")
	for {
		events, err := feed.EventsFromRSS(ctx, regionIDs)
		if err != nil {
			return err
		}
		fmt.Printf("Events: %d\n", len(events))
		time.Sleep(time.Second * 5)
	}

	return nil
	// feed.CollectEvents(regions)

	// // Validate regions provided in config
	// collector, err := feed.NewCollector(
	// 	strings.Split(conf.Regions, ","),
	// 	conf.RSSBaseURL,
	// 	time.Second*1,
	// 	time.Second*60,
	// )
	// if err != nil {
	// 	return err
	// }

	// receive, err := collector.Subscribe(ctx)
	// if err != nil {
	// 	return err
	// }
	// for {
	// 	evt, err := receive()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(evt)
	// }
}
