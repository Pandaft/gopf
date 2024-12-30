package forwarder

import (
	"fmt"
	"gopf/config"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Forwarder struct {
	rule        *config.ForwardRule
	listener    net.Listener
	done        chan struct{}
	mu          sync.Mutex
	active      sync.WaitGroup
	bytesSent   uint64
	bytesRecv   uint64
	connections uint64
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
	atomic.StoreUint64(&f.bytesSent, 0)
	atomic.StoreUint64(&f.bytesRecv, 0)
	atomic.StoreUint64(&f.connections, 0)
	go f.accept()
	go f.updateStats()
	return nil
}

func (f *Forwarder) updateStats() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-f.done:
			return
		case <-ticker.C:
			atomic.StoreUint64(&f.rule.BytesSent, atomic.LoadUint64(&f.bytesSent))
			atomic.StoreUint64(&f.rule.BytesRecv, atomic.LoadUint64(&f.bytesRecv))
			atomic.StoreUint64(&f.rule.Connections, atomic.LoadUint64(&f.connections))
		}
	}
}

func (f *Forwarder) Stop() {
	f.mu.Lock()
	if f.listener != nil {
		close(f.done)
		f.listener.Close()
		f.listener = nil
	}
	f.mu.Unlock()
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

		atomic.AddUint64(&f.connections, 1)
		go f.handleConnection(conn)
	}
}

func (f *Forwarder) handleConnection(local net.Conn) {
	defer atomic.AddUint64(&f.connections, ^uint64(0))

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
		select {
		case <-f.done:
			return
		default:
			n, err := src.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				if _, err := dst.Write(buf[:n]); err != nil {
					return
				}

				if src.LocalAddr().String() == fmt.Sprintf(":%d", f.rule.LocalPort) {
					atomic.AddUint64(&f.bytesSent, uint64(n))
				} else {
					atomic.AddUint64(&f.bytesRecv, uint64(n))
				}
			}
		}
	}
}

func (f *Forwarder) ClearStats() {
	atomic.StoreUint64(&f.bytesSent, 0)
	atomic.StoreUint64(&f.bytesRecv, 0)
	atomic.StoreUint64(&f.connections, 0)
	atomic.StoreUint64(&f.rule.BytesSent, 0)
	atomic.StoreUint64(&f.rule.BytesRecv, 0)
	atomic.StoreUint64(&f.rule.Connections, 0)
}
