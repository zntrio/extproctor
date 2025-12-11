// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

// Client wraps the ExtProc gRPC client.
type Client struct {
	conn   *grpc.ClientConn
	client extprocv3.ExternalProcessorClient
	target string
}

// Option configures the client.
type Option func(*clientConfig)

type clientConfig struct {
	target     string
	unixSocket string
	tls        bool
	tlsCert    string
	tlsKey     string
	tlsCA      string
}

// WithTarget sets the target address.
func WithTarget(target string) Option {
	return func(c *clientConfig) {
		c.target = target
	}
}

// WithUnixSocket sets the Unix domain socket path for the connection.
// When set, this takes precedence over the TCP target address.
func WithUnixSocket(path string) Option {
	return func(c *clientConfig) {
		c.unixSocket = path
	}
}

// WithTLS enables TLS with the given certificate files.
func WithTLS(cert, key, ca string) Option {
	return func(c *clientConfig) {
		c.tls = true
		c.tlsCert = cert
		c.tlsKey = key
		c.tlsCA = ca
	}
}

// New creates a new ExtProc client.
func New(opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		target: "localhost:50051",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var dialOpts []grpc.DialOption

	// Determine the connection target
	target := cfg.target
	if cfg.unixSocket != "" {
		// Use Unix domain socket - format: unix:///path/to/socket
		target = "unix://" + cfg.unixSocket
		// TLS is typically not used with Unix sockets
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else if cfg.tls {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn:   conn,
		client: extprocv3.NewExternalProcessorClient(conn),
		target: target,
	}, nil
}

// buildTLSConfig creates a TLS configuration from the provided files.
func buildTLSConfig(cfg *clientConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.tlsCert != "" && cfg.tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.tlsCert, cfg.tlsKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.tlsCA != "" {
		caCert, err := os.ReadFile(cfg.tlsCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ProcessingResult contains the responses from an ExtProc processing session.
type ProcessingResult struct {
	Responses []*PhaseResponse
}

// PhaseResponse represents a response for a specific processing phase.
type PhaseResponse struct {
	Phase    extproctorv1.ProcessingPhase
	Response *extprocv3.ProcessingResponse
}

// Process executes an ExtProc session with the given HTTP request definition.
func (c *Client) Process(ctx context.Context, req *extproctorv1.HttpRequest) (*ProcessingResult, error) {
	stream, err := c.client.Process(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start processing stream: %w", err)
	}

	result := &ProcessingResult{}

	// Send request headers
	headersReq := buildRequestHeaders(req)
	if err := stream.Send(headersReq); err != nil {
		return nil, fmt.Errorf("failed to send request headers: %w", err)
	}

	// Receive response for request headers
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response for request headers: %w", err)
	}
	result.Responses = append(result.Responses, &PhaseResponse{
		Phase:    extproctorv1.ProcessingPhase_REQUEST_HEADERS,
		Response: resp,
	})

	// Check if we should continue processing
	if isImmediateResponse(resp) {
		return result, stream.CloseSend()
	}

	// Send request body if configured
	if req.ProcessRequestBody && len(req.Body) > 0 {
		bodyReq := buildRequestBody(req)
		if err := stream.Send(bodyReq); err != nil {
			return nil, fmt.Errorf("failed to send request body: %w", err)
		}

		resp, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("failed to receive response for request body: %w", err)
		}
		result.Responses = append(result.Responses, &PhaseResponse{
			Phase:    extproctorv1.ProcessingPhase_REQUEST_BODY,
			Response: resp,
		})

		if isImmediateResponse(resp) {
			return result, stream.CloseSend()
		}
	}

	// Send request trailers if configured
	if req.ProcessRequestTrailers && len(req.Trailers) > 0 {
		trailersReq := buildRequestTrailers(req)
		if err := stream.Send(trailersReq); err != nil {
			return nil, fmt.Errorf("failed to send request trailers: %w", err)
		}

		resp, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("failed to receive response for request trailers: %w", err)
		}
		result.Responses = append(result.Responses, &PhaseResponse{
			Phase:    extproctorv1.ProcessingPhase_REQUEST_TRAILERS,
			Response: resp,
		})
	}

	return result, stream.CloseSend()
}

// isImmediateResponse checks if the response is an immediate response (short-circuit).
func isImmediateResponse(resp *extprocv3.ProcessingResponse) bool {
	return resp.GetImmediateResponse() != nil
}

// buildRequestHeaders creates a ProcessingRequest for request headers.
func buildRequestHeaders(req *extproctorv1.HttpRequest) *extprocv3.ProcessingRequest {
	headers := make([]*corev3.HeaderValue, 0, len(req.Headers)+4)

	// Add pseudo-headers
	headers = append(headers,
		&corev3.HeaderValue{Key: ":method", Value: req.Method},
		&corev3.HeaderValue{Key: ":path", Value: req.Path},
	)

	if req.Scheme != "" {
		headers = append(headers, &corev3.HeaderValue{Key: ":scheme", Value: req.Scheme})
	}

	if req.Authority != "" {
		headers = append(headers, &corev3.HeaderValue{Key: ":authority", Value: req.Authority})
	}

	// Add regular headers
	for k, v := range req.Headers {
		headers = append(headers, &corev3.HeaderValue{Key: k, Value: v})
	}

	return &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{
					Headers: headers,
				},
				EndOfStream: !req.ProcessRequestBody && !req.ProcessRequestTrailers,
			},
		},
	}
}

// buildRequestBody creates a ProcessingRequest for the request body.
func buildRequestBody(req *extproctorv1.HttpRequest) *extprocv3.ProcessingRequest {
	return &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestBody{
			RequestBody: &extprocv3.HttpBody{
				Body:        req.Body,
				EndOfStream: !req.ProcessRequestTrailers,
			},
		},
	}
}

// buildRequestTrailers creates a ProcessingRequest for request trailers.
func buildRequestTrailers(req *extproctorv1.HttpRequest) *extprocv3.ProcessingRequest {
	trailers := make([]*corev3.HeaderValue, 0, len(req.Trailers))
	for k, v := range req.Trailers {
		trailers = append(trailers, &corev3.HeaderValue{Key: k, Value: v})
	}

	return &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestTrailers{
			RequestTrailers: &extprocv3.HttpTrailers{
				Trailers: &corev3.HeaderMap{
					Headers: trailers,
				},
			},
		},
	}
}

// Target returns the target address of the client.
func (c *Client) Target() string {
	return c.target
}
