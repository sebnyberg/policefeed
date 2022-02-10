package policefeed

import "flag"

var (
	rssBaseURL  = flag.String("rss-url-template", "https://polisen.se/aktuellt/rss/norrbotten/handelser-rss---%v/", "RSS template URL, where %v is substituted with region ids")
	regionIDs   = flag.String("region-ids", "stockholms-lan", "comma-separated list of region ids (for use in the RSS URL)")
	regionNames = flag.String("region-names", "Stockholms Län", "comma-separated list of region names")
)
