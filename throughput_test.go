package benchmate

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	o := DefaultThroughputOptions()
	o.MsgSize = 64000
	o.Addr = fmt.Sprintf(":%d", randPort())
	o.ClientPort = randPort()

	l, err := net.Listen(o.Network, o.Addr)
	if err != nil {
		t.Errorf("Error making listener: %v", err)
	}

	go func() {
		err := o.ThroughputServer().Run(l)
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(time.Second)

	conn, err := net.Dial(o.Network, o.Addr)
	if err != nil {
		t.Errorf("Error making connection: %v", err)
	}

	result, err := o.ThroughputClient().Run(conn)
	if err != nil {
		t.Errorf("Error running throughput test: %v", err)
	}

	fmt.Println("average throughput:", float64(result.NumMsg*result.MsgSize)*1000/float64(result.Elapsed.Nanoseconds()), "MB/s")

	t.Log(result)

	time.Sleep(time.Millisecond * 500)
}
