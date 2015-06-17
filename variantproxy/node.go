package variantproxy

import (
	"github.com/foomo/variant-balancer/config"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Node struct {
	Server            string
	Url               *url.URL
	SessionCookieName string
	Id                string
	openConnections   int
	maxConnections    int
	ReverseProxy      *httputil.ReverseProxy
	channelOpenConn   chan int
	channelCloseConn  chan int
}

func NewNode(nodeConfig *config.Node) *Node {
	url, err := url.Parse(nodeConfig.Server)
	if err != nil {
		panic(err)
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	n := &Node{
		Server:            nodeConfig.Server,
		Url:               url,
		Id:                nodeConfig.Id,
		ReverseProxy:      reverseProxy,
		SessionCookieName: nodeConfig.Cookie,
		openConnections:   0,
		maxConnections:    nodeConfig.MaxConnections,
		channelOpenConn:   make(chan int),
		channelCloseConn:  make(chan int),
	}
	go func() {
		for {
			select {
			case <-n.channelCloseConn:
				n.openConnections--
			case <-n.channelOpenConn:
				n.openConnections++
			}
		}
	}()
	return n
}

func (n *Node) Load() float64 {
	if n.openConnections > 0 {
		l := float64(n.openConnections) / float64(n.maxConnections)
		return l
	} else {
		return 0
	}
}

func (n *Node) closeConn() {
	n.channelCloseConn <- 1
}

func (n *Node) ServeHTTP(w http.ResponseWriter, incomingRequest *http.Request) {
	n.channelOpenConn <- 1
	defer func() {
		if r := recover(); r != nil {
			n.closeConn()
		}
	}()
	// there is no error propagation here
	n.ReverseProxy.ServeHTTP(w, incomingRequest)
	n.closeConn()
}
