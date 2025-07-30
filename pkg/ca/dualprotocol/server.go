// SPDX-License-Identifier: CC0-1.0

package dualprotocol

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// ConnectionInfo holds details about the connection
type ConnectionInfo struct {
	Protocol    string
	IsTLS       bool
	TLSVersion  string
	CipherSuite string
	RemoteAddr  string
	DetectedAt  time.Time
}

// contextKey is used for storing connection info in request context
type contextKey string

const (
	// ConnectionInfoKey is the key for storing ConnectionInfo in request context
	ConnectionInfoKey contextKey = "dual-protocol-connection-info"

	// TLS record type for handshake (first byte of TLS connection)
	tlsHandshakeType = 0x16

	// Detection timeout for protocol detection
	detectionTimeout = 5 * time.Second

	// Buffer size for peeking at connection data
	peekBufferSize = 1
)

// GetConnectionInfo retrieves connection information from the request context
func GetConnectionInfo(r *http.Request) (*ConnectionInfo, bool) {
	info, ok := r.Context().Value(ConnectionInfoKey).(*ConnectionInfo)
	return info, ok
}

// Server provides a server that can handle both HTTP and HTTPS
// connections on the same port by detecting the protocol from connection bytes
type Server struct {
	*http.Server
	tlsConfig      *tls.Config
	listener       net.Listener
	dualListener   *dualListener
	shutdownOnce   sync.Once
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	logger         Logger
}

// Logger interface for logging operations
type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
}

// dualListener wraps a net.Listener to detect HTTP vs HTTPS protocol
type dualListener struct {
	net.Listener
	tlsConfig *tls.Config
	logger    Logger
}

// dualConn wraps net.Conn to detect TLS vs HTTP protocol
type dualConn struct {
	net.Conn
	originalConn net.Conn // Store original connection to avoid circular references
	reader       *bufio.Reader
	tlsConfig    *tls.Config
	detected     bool
	isTLS        bool
	buffer       []byte
	connInfo     *ConnectionInfo
	logger       Logger
}

// bufferedConn wraps a connection to use a buffered reader for reads
// while delegating other operations to the underlying connection
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

// Read reads data from the buffered reader
func (bc *bufferedConn) Read(b []byte) (int, error) {
	return bc.reader.Read(b)
}

// GetConnectionInfo returns the connection info for this connection
func (c *dualConn) GetConnectionInfo() *ConnectionInfo {
	return c.connInfo
}

// WrapHandlerWithConnectionInfo wraps an HTTP handler to inject connection info into the request context
func WrapHandlerWithConnectionInfo(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create connection info based on the request
		connInfo := &ConnectionInfo{
			Protocol:   "HTTP",
			IsTLS:      r.TLS != nil,
			RemoteAddr: r.RemoteAddr,
			DetectedAt: time.Now(),
		}

		if r.TLS != nil {
			connInfo.Protocol = "HTTPS"
			connInfo.TLSVersion = getTLSVersionString(r.TLS.Version)
			connInfo.CipherSuite = tls.CipherSuiteName(r.TLS.CipherSuite)
		}

		// Add connection info to request context
		ctx := context.WithValue(r.Context(), ConnectionInfoKey, connInfo)
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	})
}

// NewServer creates a new server that can handle both HTTP and HTTPS
// connections on the same port
func NewServer(server *http.Server, tlsConfig *tls.Config, logger Logger) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		Server:         server,
		tlsConfig:      tlsConfig,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
		logger:         logger,
	}
}

// ListenAndServe starts the dual protocol server on the configured address
func (s *Server) ListenAndServe() error {
	if s.Addr == "" {
		s.Addr = ":8443"
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Addr, err)
	}

	s.listener = listener
	s.dualListener = &dualListener{
		Listener:  listener,
		tlsConfig: s.tlsConfig,
		logger:    s.logger,
	}

	s.logger.Info("Starting dual protocol server", "addr", s.Addr)

	return s.Server.Serve(s.dualListener)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	var err error
	s.shutdownOnce.Do(func() {
		s.shutdownCancel()
		if s.Server != nil {
			err = s.Server.Shutdown(ctx)
		}
	})
	return err
}

// Accept implements net.Listener interface with protocol detection
func (l *dualListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &dualConn{
		Conn:         conn,
		originalConn: conn, // Store original connection
		reader:       bufio.NewReader(conn),
		tlsConfig:    l.tlsConfig,
		logger:       l.logger,
		connInfo: &ConnectionInfo{
			RemoteAddr: conn.RemoteAddr().String(),
			DetectedAt: time.Now(),
		},
	}, nil
}

// Read implements io.Reader with protocol detection on first read
func (c *dualConn) Read(b []byte) (int, error) {
	if !c.detected {
		if err := c.detectProtocol(); err != nil {
			return 0, fmt.Errorf("protocol detection failed: %w", err)
		}
	}

	// If we have buffered data, read from buffer first
	if len(c.buffer) > 0 {
		n := copy(b, c.buffer)
		c.buffer = c.buffer[n:]
		return n, nil
	}

	return c.reader.Read(b)
}

// detectProtocol determines if the connection is HTTP or HTTPS
func (c *dualConn) detectProtocol() error {
	c.detected = true

	// Set detection timeout
	if err := c.Conn.SetReadDeadline(time.Now().Add(detectionTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Peek at the first byte to determine protocol
	firstByte, err := c.reader.Peek(peekBufferSize)
	if err != nil {
		return fmt.Errorf("failed to peek at connection data: %w", err)
	}

	// Reset deadline after detection
	if err := c.Conn.SetReadDeadline(time.Time{}); err != nil {
		c.logger.Warn("Failed to reset read deadline", "error", err)
	}

	// Check if it's a TLS handshake
	if len(firstByte) > 0 && firstByte[0] == tlsHandshakeType {
		c.isTLS = true
		c.connInfo.IsTLS = true
		c.connInfo.Protocol = "HTTPS"
		c.logger.Debug("Detected TLS connection", "remote_addr", c.Conn.RemoteAddr())
		return c.upgradeToTLS()
	}

	c.isTLS = false
	c.connInfo.IsTLS = false
	c.connInfo.Protocol = "HTTP"
	c.logger.Debug("Detected HTTP connection", "remote_addr", c.Conn.RemoteAddr())
	return nil
}

// upgradeToTLS upgrades the connection to use TLS
func (c *dualConn) upgradeToTLS() error {
	if c.tlsConfig == nil {
		return fmt.Errorf("TLS config not provided for TLS connection")
	}

	// Create a connection wrapper that uses the buffered reader for reads
	// but delegates other operations to the original connection
	wrappedConn := &bufferedConn{
		Conn:   c.originalConn,
		reader: c.reader,
	}

	// Create TLS server connection using the wrapped connection
	tlsConn := tls.Server(wrappedConn, c.tlsConfig)

	// Perform TLS handshake with timeout
	handshakeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	handshakeChan := make(chan error, 1)
	go func() {
		handshakeChan <- tlsConn.Handshake()
	}()

	select {
	case err := <-handshakeChan:
		if err != nil {
			return fmt.Errorf("TLS handshake failed: %w", err)
		}
	case <-handshakeCtx.Done():
		return fmt.Errorf("TLS handshake timeout")
	}

	// Update connection info with TLS details
	state := tlsConn.ConnectionState()
	c.connInfo.TLSVersion = getTLSVersionString(state.Version)
	c.connInfo.CipherSuite = tls.CipherSuiteName(state.CipherSuite)

	// Replace the reader with TLS connection
	c.reader = bufio.NewReader(tlsConn)
	c.Conn = tlsConn

	c.logger.Debug("TLS handshake completed",
		"remote_addr", c.Conn.RemoteAddr(),
		"tls_version", c.connInfo.TLSVersion,
		"cipher_suite", c.connInfo.CipherSuite)

	return nil
}

// getTLSVersionString converts TLS version constant to string using built-in Go constants
func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown TLS version: %x", version)
	}
}
