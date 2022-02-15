# Kobo exporter

A [Prometheus exporter] for [Kobo store] prices.

[Prometheus exporter]: https://prometheus.io/docs/instrumenting/exporters/
[Kobo store]: https://www.kobo.com/gb/en/ebooks

## Set up

```shell
asdf install
make build
```

## Running

`./kobo_exporter` serves a Prometheus exporter at `/metrics`. It takes
the following command-line arguments:

- `-port` - port to serve on. Defaults to 8080.
- `-frequency` - scrape frequency for Kobo store prices, in seconds.
  Each book is scraped in turn, so the initial prices will be complete
  after `frequency * len(books)`. Defaults to 600 (10 minutes).
- `-config` - config file. Each line is a URL to scrape. See
  [`urls.conf.example`](urls.conf.example) for a small example. Defaults
  to empty (no URLs to scrape).
- Any additional positional arguments are treated as URLs to scrape,
  assuming `-config` is not present. That is, both of these are valid:
    - `./kobo_exporter -config my_urls.conf`
    - `./kobo_exporter https://www.kobo.com/... https://www.kobo.com/...`
