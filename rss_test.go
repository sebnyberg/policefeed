package policefeed_test

import (
	"os"
	"testing"

	policefeed "github.com/sebnyberg/police-feed-se"
	"github.com/stretchr/testify/require"
)

func TestParseRSS(t *testing.T) {
	f, err := os.OpenFile("testdata/example-rss.xml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	res, err := policefeed.EventsFromRSS(f)
	require.NoError(t, err)
	require.NotNil(t, res)
}
