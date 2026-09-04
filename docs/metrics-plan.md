# Extended plan: parameterized metrics support

Written by Sara for Ivan, 2026-09-04. Planning only — no code changes in this doc's
commit. Researched by cloning `Altinity/altinity-mcp` read-only and reading the source;
nothing below is guessed.

## What's already there

* `pkg/server/catalog_cache.go` already has a `CatalogCacheMetrics` struct with atomic
  counters: `HitsOK`, `HitsDenied`, `Misses`, `FullDropsOK`, `FullDropsDenied`,
  `DiscoverySaturate`, `DiscoveryAuthErr`, `DiscoveryTransErr`. These are real,
  in-process, thread-safe counters — they are simply never exposed outside the process.
  The comment above the struct says as much: "the cache layer has no metrics-backend
  dependency."
* `cmd/altinity-mcp/main.go` builds an `http.NewServeMux()` per transport mode
  (streamable-http, SSE, and others — the pattern repeats ~4 times in the file), and
  every one of them registers `/health`, `/livez`, and `/jwe-token-generator` on that
  mux. A `/metrics` endpoint is the same shape of change, in the same place, once per
  transport block.
* No metrics/observability dependency (Prometheus client, OpenTelemetry, etc.) is
  imported anywhere in the module today — this is a clean addition, not a migration.

## Recommended approach

1. Add `github.com/prometheus/client_golang/prometheus` (+ `promhttp`) as a dependency
   — the standard, idiomatic choice for a Go service exposing metrics, and it plugs
   directly into the existing `http.ServeMux` pattern via `promhttp.Handler()`.
2. Register `/metrics` alongside `/health` and `/livez` in each of the mux-building
   blocks in `main.go`, gated by a config flag (`Server.Metrics.Enabled` or similar) so
   it can be turned off for anyone who doesn't want it — matches how TLS and OAuth are
   already optional in `pkg/config`.
3. Wrap `CatalogCacheMetrics` in a Prometheus `Collector` (or convert the atomic
   counters to `prometheus.Counter`s directly) so the existing instrumentation becomes
   visible for free — no new counting logic, just a new consumer of numbers already
   being counted.

## New metrics worth adding

All of these sit at points in `pkg/clickhouse/client.go` and `pkg/server/server_query.go`
that don't currently record anything:

* **Query count and latency**, labeled by outcome (`ok` / `error`) and, since the
  server supports multiple ClickHouse clusters (`pkg/server/multicluster_router.go`,
  `multicluster_factory.go`), labeled by cluster too. Natural insertion points:
  `Client.ExecuteQuery`, `ExecuteCappedQuery`, `executeSelect`, `executeNonSelect`.
* **Rows / bytes returned per query** — `ExecuteCappedQuery` already tracks `maxRows`
  and `maxBytes` caps; the actual counts are a small addition at the same call site.
* **Blocked-clause rejections** — `server_query.go`'s `checkBlockedClauses` is a
  security guard (denies specific SQL clauses); a counter here gives visibility into
  how often the guard actually fires, which today is invisible.
* **ClickHouse connection health** — `Client.Ping` already exists; a gauge fed from it
  turns an ad hoc health check into a scraped signal.
* **Catalog cache** (see above) — hits/misses/discovery-errors, exposed as-is.

## Deliberately not planned yet

* No OpenTelemetry tracing — metrics only, per what was asked. Tracing is a much
  bigger surface (context propagation through every handler) and wasn't requested.
* No Helm chart changes (`helm/altinity-mcp/`) for a `ServiceMonitor` or Prometheus
  scrape annotations in this pass — worth a follow-up once the metrics themselves
  exist and their names are stable enough to commit to in a chart.
* No decision here on cardinality limits for cluster labels (a router with many
  clusters could produce a lot of series) — flagging it so the implementation doesn't
  skip it, not solving it in a planning doc.

## Next step

Implementation, once Ivan confirms this direction — new task per his instruction, not
bundled into this plan.
