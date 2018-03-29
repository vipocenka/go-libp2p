package tcpreuse

import (
	"net"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
)

func TestSingle(t *testing.T) {
	var trA Transport
	var trB Transport
	laddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	listenerA, err := trA.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerA.Close()
	listenerB, err := trB.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerB.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := listenerA.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}()

	c, err := trB.Dial(listenerA.Multiaddr())
	if err != nil {
		t.Fatal(err)
	}
	<-done
	c.Close()
}

func TestTwoLocal(t *testing.T) {
	var trA Transport
	var trB Transport
	laddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	listenerA, err := trA.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerA.Close()

	listenerB1, err := trB.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerB1.Close()

	listenerB2, err := trB.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerB2.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := listenerA.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}()

	c, err := trB.Dial(listenerA.Multiaddr())
	if err != nil {
		t.Fatal(err)
	}
	localPort := c.LocalAddr().(*net.TCPAddr).Port
	if localPort != listenerB1.Addr().(*net.TCPAddr).Port &&
		localPort != listenerB2.Addr().(*net.TCPAddr).Port {
		t.Fatal("didn't dial from one of our listener ports")
	}
	<-done
	c.Close()
}

func TestLocalAndUnspecified(t *testing.T) {
	var trA Transport
	var trB Transport
	laddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	unspec, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	listenerA, err := trA.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerA.Close()

	listenerB1, err := trB.Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerB1.Close()

	listenerB2, err := trB.Listen(unspec)
	if err != nil {
		t.Fatal(err)
	}
	defer listenerB2.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := listenerA.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}()

	c, err := trB.Dial(listenerA.Multiaddr())
	if err != nil {
		t.Fatal(err)
	}
	actual := c.LocalAddr().(*net.TCPAddr).Port
	expected := listenerB1.Addr().(*net.TCPAddr).Port
	if actual != expected {
		t.Errorf("expected to use port %d, used port %d", expected, actual)
	}
	<-done
	c.Close()

	// Closing the listener should reset the dialer.
	listenerB1.Close()

	done = make(chan struct{})
	go func() {
		defer close(done)
		c, err := listenerA.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}()

	c, err = trB.Dial(listenerA.Multiaddr())
	if err != nil {
		t.Fatal(err)
	}
	actual = c.LocalAddr().(*net.TCPAddr).Port
	expected = listenerB2.Addr().(*net.TCPAddr).Port
	if actual != expected {
		t.Errorf("expected to use port %d, used port %d", expected, actual)
	}
	<-done
	c.Close()
}
