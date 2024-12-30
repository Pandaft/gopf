package forwarder

import (
	"fmt"
	"gopf/config"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type Forwarder struct {
	rule     *config.ForwardRule
	listener net.Listener
	done     chan struct{}
	mu       sync.Mutex
}

func NewForwarder(rule *config.ForwardRule) *Forwarder {
	return &Forwarder{
		rule: rule,
		done: make(chan struct{}),
	}
}

func (f *Forwarder) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", f.rule.LocalPort))
	if err != nil {
		return err
	}

	f.listener = listener
	go f.accept()
	return nil
}

func (f *Forwarder) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.listener != nil {
		f.listener.Close()
		close(f.done)
	}
}

func (f *Forwarder) accept() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			select {
			case <-f.done:
				return
			default:
				continue
			}
		}

		atomic.AddUint64(&f.rule.Connections, 1)
		go f.handleConnection(conn)
	}
}

func (f *Forwarder) handleConnection(local net.Conn) {
	remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", f.rule.RemoteHost, f.rule.RemotePort))
	if err != nil {
		local.Close()
		return
	}

	go f.pipe(local, remote)
	go f.pipe(remote, local)
}

func (f *Forwarder) pipe(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				return
			}
			return
		}

		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return
			}

			if src.LocalAddr().String() == fmt.Sprintf(":%d", f.rule.LocalPort) {
				atomic.AddUint64(&f.rule.BytesSent, uint64(n))
			} else {
				atomic.AddUint64(&f.rule.BytesRecv, uint64(n))
			}
		}
	}
}
