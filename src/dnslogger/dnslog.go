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
	Ch            ch.Conn
	Table         string
	ResponseTable string
}

func (svc *DNSLogServiceServer) Init(clickhouse string, u string, p string, d string, t string, rt string, w int) {
	Cho, err := ch.Open(&ch.Options{
		Addr: strings.Split(clickhouse, ","),
		Auth: ch.Auth{
			Database: d,
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
	svc.Table = t
	svc.ResponseTable = rt
}

func (svc *DNSLogServiceServer) Worker(conn net.Conn) error {
	b := make([]byte, 1024)
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
		slog.Debugf("Read %d", n)
		if uint(n) != l+2 {
			slog.Debugf("Read %d, expect %d", n, l)
			continue
		}
		if err = proto.Unmarshal(b[2:], msg); err != nil {
			qstring := ""
			if msg.From != nil && msg.TimeSec != nil && msg.Question != nil {
				slog.Debugf("Query %s", *msg.Question.QName)
				qstring = fmt.Sprintf("INSERT INTO %s VALUES (%d,'%s','%s')", svc.Table, time.Now().Unix(), net.IP(msg.From).String(), *msg.Question.QName)
				err = svc.Ch.AsyncInsert(context.Background(), qstring, false)
				if err != nil {
					slog.Debug(err)
				}
			} else if msg.To != nil && msg.TimeSec != nil && msg.Response != nil && svc.ResponseTable != "" {
				slog.Debugf("Response %s", net.IP(msg.To).String())
				for _, rs := range msg.Response.Rrs {
					data := ""
					if *rs.Type == 1 || *rs.Type == 28 {
						data = net.IP(rs.Rdata).String()
					} else {
						data = string(data)
					}
					qstring = fmt.Sprintf("INSERT INTO %s VALUES (%d,'%s','%s',%d,'%s')", svc.ResponseTable, time.Now().Unix(), net.IP(msg.To).String(), *rs.Name, *rs.Type, data)
					err = svc.Ch.AsyncInsert(context.Background(), qstring, false)
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
