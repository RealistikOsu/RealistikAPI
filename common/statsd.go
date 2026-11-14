package common

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const statsdResolveEvery = time.Minute

// Statsd is a minimal DogStatsD sender. It does nothing while the host is empty and never returns an
// error, so metrics can't take the API down. The agent's address is looked up again every minute,
// because the agent container usually starts after this service and can change address.
type Statsd struct {
	host string
	port int
	conn *net.UDPConn

	mu          sync.Mutex
	addr        *net.UDPAddr
	nextResolve time.Time
}

func NewStatsd(host string, port int) *Statsd {
	s := &Statsd{host: host, port: port}
	if host != "" {
		s.conn, _ = net.ListenUDP("udp4", nil)
	}
	return s
}

func (s *Statsd) address() *net.UDPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Now().Before(s.nextResolve) {
		return s.addr
	}

	s.nextResolve = time.Now().Add(statsdResolveEvery)
	// Not up yet leaves the last address in place.
	if addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", s.host, s.port)); err == nil {
		s.addr = addr
	}
	return s.addr
}

func (s *Statsd) send(name string, value string, kind string, tags []string) {
	if s.conn == nil {
		return
	}

	addr := s.address()
	if addr == nil {
		return
	}

	line := fmt.Sprintf("api.%s:%s|%s", name, value, kind)
	if len(tags) > 0 {
		line += "|#" + strings.Join(tags, ",")
	}
	s.conn.WriteToUDP([]byte(line), addr)
}

func (s *Statsd) Incr(name string, tags []string) {
	s.send(name, "1", "c", tags)
}

func (s *Statsd) Timing(name string, d time.Duration, tags []string) {
	s.send(name, fmt.Sprintf("%.3f", float64(d)/float64(time.Millisecond)), "ms", tags)
}
