# Swedish Police Feed

Police events service based on [Swedish RSS feeds](https://polisen.se/aktuellt/rss/lokala-rss-floden/).

## Running the server

Install the binary:

```bash
go install github.com/sebnyberg/policefeed
```

The service uses Postgres for persistence. Provide Postgres credentials either as CLI flags, or as [Postgres Environment Variables](https://www.postgresql.org/docs/9.3/libpq-envars.html).

For CLI options, run:

```bash
policefeed server --help
```

Example running Postgres on localhost:

```bash
policefeed server \
  --addr localhost:8080 \
  --pgdb dev \
  --pgpassword secretpassword \
  --pghost localhost \
  --pgport 5432 \
  --pgsslmode disable \
  --pguser myuser
```

## Development

Start the Postgres database with Docker-Compose

```bash
docker-compose up -d
```

Run the server

```bash
go run main.go server
```

## Legal considerations

As per the [Police website](https://polisen.se/aktuellt/rss/):

> Du får använda våra RSS-flöden på din egen webbplats under förutsättning att det tydligt framgår att polisen är källan och att du länkar till artikeln på polisen.se. Det är inte tillåtet att använda polisens logotyp.

Or in English:

> You may use our RSS-feeds in your own application given that it is evidently clear that the police is the source, and that you link to the event's corresponding article at polisen.se. It is forbidden to use the police logo.

As per the MIT license, I waive any liability in the event that you misuse this software. The default configuration *should* be conservative enough to avoid undue pressure on the Police's servers.
