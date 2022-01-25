package benchmate

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestLatency(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	o := DefaultLatencyOptions()
	o.MsgSize = 128
	o.Addr = fmt.Sprintf(":%d", randPort())
	o.ClientPort = randPort()

	l, err := net.Listen(o.Network, o.Addr)
	if err != nil {
		t.Errorf("Error making listener: %v", err)
	}

	go func() {
		_ = o.LatencyServer().Run(l)
	}()

	time.Sleep(time.Second)

	conn, err := net.Dial(o.Network, o.Addr)
	if err != nil {
		t.Errorf("Error making connection: %v", err)
	}

	result, err := o.LatencyClient().Run(conn)
	if err != nil {
		t.Errorf("Error running throughput test: %v", err)
	}

	fmt.Println("average latency:", time.Duration(float64(result.ElapsedTime.Nanoseconds())/float64(result.NumMsg)))

	t.Log(result)

	time.Sleep(time.Millisecond * 500)
}
