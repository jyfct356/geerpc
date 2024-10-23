package timeout

import (
	"fmt"
	"geerpc/client"
	"geerpc/server"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestClient_dialTimeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")

	f := func(conn net.Conn, opt *server.Option) (client *client.Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}
	t.Run("timeout", func(t *testing.T) {
		_, err := client.DialTimeout(f, "tcp", l.Addr().String(), &server.Option{ConnectTimeout: time.Second})
		_assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "expect a timeout error")
	})
	t.Run("0", func(t *testing.T) {
		_, err := client.DialTimeout(f, "tcp", l.Addr().String(), &server.Option{ConnectTimeout: 0})
		_assert(err == nil, "0 means no limit")
	})
}

func TestXDial(t *testing.T) {
	defaultServer := server.DefaultServer
	if runtime.GOOS == "linux" {
		ch := make(chan struct{})
		addr := "/tmp/geerpc.sock"
		go func() {
			_ = os.Remove(addr)
			l, err := net.Listen("unix", addr)
			if err != nil {
				t.Fatal("failed to listen unix socket")
			}
			ch <- struct{}{}
			defaultServer.Accept(l)
		}()
		<-ch
		_, err := client.XDial("unix@" + addr)
		_assert(err == nil, "failed to connect unix socket")
	}
}
