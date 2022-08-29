package dnslogger

import (
	"context"
	"dnsmessage"
	"fmt"
	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang/protobuf/proto"
	slog "github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	slog.Infof(format, args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	slog.Errorf(format, args...)
}

type DNSLogServiceServer struct {
	Ch ch.Conn
}

func (svc *DNSLogServiceServer) Init(clickhouse string, u string, p string, w int) {
	Cho, err := ch.Open(&ch.Options{
		Addr: strings.Split(clickhouse, ","),
		Auth: ch.Auth{
			Database: "pdns",
			Username: u,
			Password: p,
		},
		DialTimeout:     time.Second,
		MaxOpenConns:    w,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		slog.Fatalf("Error creating the client: %s", err)
	}
	svc.Ch = Cho
}

func (svc *DNSLogServiceServer) Worker(conn net.Conn) error {
	b := make([]byte,1024)
	var l uint
	msg := &dnsmessage.PBDNSMessage{}
	slog.Debug("Worker")
	for {
		n, err := conn.Read(b)
		if err != nil {
			slog.Debug("Error reading")
			break
		}
		l = uint(b[1]) + uint(b[0])*256
		slog.Debug("Read successful")
		slog.Debugf("Read %d",n)
		if uint(n) != l+2 {
			slog.Debugf("Read %d, expect %d",n, l)
			continue
		}
		if err = proto.Unmarshal(b[2:], msg); err != nil {
			if msg.From != nil && msg.TimeSec != nil && msg.Question != nil {
				slog.Debugf("Query %s",*msg.Question.QName)
				qstring := fmt.Sprintf("INSERT INTO pdns_query_logs VALUES (%d,'%s','%s')", time.Now().Unix(), net.IP(msg.From).String(), *msg.Question.QName)
				err = svc.Ch.AsyncInsert(context.Background(), qstring, false)
				if err != nil {
					slog.Debug(err)
				}
			} else {
				slog.Debugf("Empty content %d", msg.Type)
			}
		} else {
			slog.Debug("Parse error")
		}
	}
	slog.Debug("Worker exited")
	_ = conn.Close()
	return nil
}
