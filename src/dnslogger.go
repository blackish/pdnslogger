package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"dnslogger"

	log "github.com/sirupsen/logrus"
)

var (
	debug bool

	port uint

	username string

	password string

	back string

	clickhouse string

	workers int

	dbname string

	tablename string

	responseTablename string
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18090, "Accesslog server port")
	flag.StringVar(&clickhouse, "clickhouse", "127.0.0.1:9000", "Clickhouse endpoints, comma separated")
	flag.StringVar(&username, "username", "", "Backend usename")
	flag.StringVar(&password, "password", "", "Backend password")
	flag.StringVar(&dbname, "database", "pdns", "Database name")
	flag.StringVar(&tablename, "table", "pdns_query_logs", "Table name")
	flag.StringVar(&responseTablename, "response_table", "", "Response table name")
	flag.IntVar(&workers, "workers", 10, "Number of open connections to clickhouse")
}

func RunLoggerServer(ctx context.Context, sl *dnslogger.DNSLogServiceServer, port uint) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Failed to bind port")
	}
	for {
		conn, err := listener.Accept()
		if err == nil {
			go sl.Worker(conn)
		}
	}
	listener.Close()
	<-ctx.Done()
}

func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	ctx := context.Background()
	log.Printf("Starting log service")

	sl := &dnslogger.DNSLogServiceServer{}
	sl.Init(clickhouse, username, password, dbname, tablename, responseTablename, workers)
	go RunLoggerServer(ctx, sl, port)
	<-ctx.Done()
}
