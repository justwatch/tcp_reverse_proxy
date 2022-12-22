package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"inet.af/tcpproxy"
	"log"
)

var cli struct {
	ListenPort    int    `kong:"default='8080',env='LISTEN_PORT',help='port to listen on'"`
	TargetAddress string `kong:"default='localhost:9200',env='TARGET_ADDRESS',help='upstream address to connect to. Can be IP or name, later one will be resolved'"`
}

func main() {
	kong.Parse(&cli)
	var p tcpproxy.Proxy
	p.AddRoute(fmt.Sprintf(":%d", cli.ListenPort), tcpproxy.To(cli.TargetAddress))
	log.Fatal(p.Run())
}
