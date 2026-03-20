// core/scanner/ping.go
package scanner

import (
	"context"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"smegg.me/thughunter/common/logger"
)

// ping tries ICMP first (if enabled), then falls back to TCP.
func ping(ctx context.Context, ip string, timeoutSec int, useICMP bool) PingResult {
	if ctx.Err() != nil {
		return PingResult{Status: StatusError, Error: "cancelled"}
	}
	if useICMP {
		result := icmpPing(ctx, ip, timeoutSec)
		if result.Status == StatusOK {
			logger.Debug().Str("ip", ip).Dur("latency", result.Latency).Msg("ICMP ping succeeded")
			return result
		}
		logger.Debug().Str("ip", ip).Str("err", result.Error).Msg("ICMP ping failed, falling back to TCP")
	}
	result := tcpPing(ctx, ip, timeoutSec)
	logger.Debug().Str("ip", ip).Str("status", string(result.Status)).Msg("TCP ping result")
	return result
}

// icmpPing sends a single ICMP echo request and waits for the reply.
func icmpPing(ctx context.Context, ip string, timeoutSec int) PingResult {
	conn, err := icmp.ListenPacket("udp4", "")
	if err != nil {
		return PingResult{Status: StatusError, Error: err.Error()}
	}
	defer conn.Close()

	// Close conn on context cancellation to unblock ReadFrom immediately.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	timeout := time.Duration(timeoutSec) * time.Second
	conn.SetDeadline(time.Now().Add(timeout))

	msg := buildEchoRequest()
	dst := &net.UDPAddr{IP: net.ParseIP(ip)}

	start := time.Now()
	if _, err := conn.WriteTo(msg, dst); err != nil {
		return PingResult{Status: StatusError, Error: err.Error()}
	}

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	latency := time.Since(start)

	if err != nil {
		return classifyPingError(err, latency)
	}

	if err := parseEchoReply(reply[:n]); err != nil {
		return PingResult{Status: StatusError, Latency: latency, Error: err.Error()}
	}

	return PingResult{Alive: true, Latency: latency, Status: StatusOK}
}

// buildEchoRequest constructs a marshalled ICMP echo request.
func buildEchoRequest() []byte {
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("thughunter-ping"),
		},
	}
	b, _ := msg.Marshal(nil)
	return b
}

// parseEchoReply validates that the response is an ICMP echo reply.
func parseEchoReply(b []byte) error {
	rm, err := icmp.ParseMessage(1, b) // protocol 1 = ICMP
	if err != nil {
		return err
	}
	if rm.Type != ipv4.ICMPTypeEchoReply {
		return net.UnknownNetworkError("unexpected ICMP type")
	}
	return nil
}

// tcpPing attempts a TCP connect to common ports as a fallback.
func tcpPing(ctx context.Context, ip string, timeoutSec int) PingResult {
	timeout := time.Duration(timeoutSec) * time.Second
	dialer := net.Dialer{Timeout: timeout}

	for _, port := range []string{"80", "443", "7"} {
		if ctx.Err() != nil {
			return PingResult{Status: StatusError, Error: "cancelled"}
		}
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
		latency := time.Since(start)

		if err == nil {
			conn.Close()
			logger.Debug().Str("ip", ip).Str("port", port).Dur("latency", latency).Msg("TCP ping connected")
			return PingResult{Alive: true, Latency: latency, Status: StatusOK}
		}
		logger.Debug().Str("ip", ip).Str("port", port).Err(err).Msg("TCP ping port unreachable")
	}

	return PingResult{Status: StatusTimeout, Error: "all TCP ping ports unreachable"}
}

// classifyPingError maps a network error into a PingResult.
func classifyPingError(err error, latency time.Duration) PingResult {
	if isTimeout(err) {
		return PingResult{Latency: latency, Status: StatusTimeout, Error: err.Error()}
	}
	return PingResult{Latency: latency, Status: StatusError, Error: err.Error()}
}

func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
