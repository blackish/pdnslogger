module main

go 1.18

require (
	dnslogger v0.0.0
	dnsmessage v0.0.0
	github.com/sirupsen/logrus v1.6.0
)

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/paulmach/orb v0.7.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/tools v0.1.4 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace dnslogger => ./src/dnslogger

replace dnsmessage => ./src/dnsmessage
