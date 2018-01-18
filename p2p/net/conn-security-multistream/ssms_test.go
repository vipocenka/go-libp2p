package ssms

import (
	"context"
	"net"
	"sync"
	"testing"

	ss "github.com/libp2p/go-stream-security"
	sst "github.com/libp2p/go-stream-security/test"
)

func TestCommonProto(t *testing.T) {
	var at, bt SSMuxer
	atInsecure := ss.InsecureTransport("peerA")
	btInsecure := ss.InsecureTransport("peerB")
	at.AddTransport("/plaintext/1.0.0", &atInsecure)
	bt.AddTransport("/plaintext/1.1.0", &btInsecure)
	bt.AddTransport("/plaintext/1.0.0", &btInsecure)
	sst.SubtestRW(t, &at, &bt, "peerA", "peerB")
}

func TestNoCommonProto(t *testing.T) {
	var at, bt SSMuxer
	atInsecure := ss.InsecureTransport("peerA")
	btInsecure := ss.InsecureTransport("peerB")

	at.AddTransport("/plaintext/1.0.0", &atInsecure)
	bt.AddTransport("/plaintext/1.1.0", &btInsecure)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, b := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer a.Close()
		_, err := at.SecureInbound(ctx, a)
		if err == nil {
			t.Fatal("conection should have failed")
		}
	}()

	go func() {
		defer wg.Done()
		defer b.Close()
		_, err := bt.SecureOutbound(ctx, b, "peerA")
		if err == nil {
			t.Fatal("connection should have failed")
		}
	}()
	wg.Wait()
}
