// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

func TestWithTarget(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithTarget("localhost:9090")
	opt(cfg)
	assert.Equal(t, "localhost:9090", cfg.target)
}

func TestWithUnixSocket(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithUnixSocket("/var/run/extproc.sock")
	opt(cfg)
	assert.Equal(t, "/var/run/extproc.sock", cfg.unixSocket)
}

func TestWithTLS(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithTLS("/path/to/cert.pem", "/path/to/key.pem", "/path/to/ca.pem")
	opt(cfg)
	assert.True(t, cfg.tls)
	assert.Equal(t, "/path/to/cert.pem", cfg.tlsCert)
	assert.Equal(t, "/path/to/key.pem", cfg.tlsKey)
	assert.Equal(t, "/path/to/ca.pem", cfg.tlsCA)
}

func TestClient_Close_NilConn(t *testing.T) {
	c := &Client{conn: nil}
	err := c.Close()
	assert.NoError(t, err)
}

func TestClient_Target(t *testing.T) {
	c := &Client{target: "localhost:50051"}
	assert.Equal(t, "localhost:50051", c.Target())
}

func TestIsImmediateResponse_True(t *testing.T) {
	resp := &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ImmediateResponse{
			ImmediateResponse: &extprocv3.ImmediateResponse{},
		},
	}
	assert.True(t, isImmediateResponse(resp))
}

func TestIsImmediateResponse_False(t *testing.T) {
	resp := &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{},
		},
	}
	assert.False(t, isImmediateResponse(resp))
}

func TestIsImmediateResponse_Nil(t *testing.T) {
	var resp *extprocv3.ProcessingResponse
	assert.False(t, isImmediateResponse(resp))

	resp = &extprocv3.ProcessingResponse{}
	assert.False(t, isImmediateResponse(resp))
}

func TestBuildRequestHeaders_Basic(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method: "GET",
		Path:   "/api/test",
	}

	procReq := buildRequestHeaders(req)
	assert.NotNil(t, procReq)

	headers := procReq.GetRequestHeaders()
	assert.NotNil(t, headers)
	assert.NotNil(t, headers.Headers)
	assert.True(t, headers.EndOfStream)

	// Check pseudo-headers
	foundMethod := false
	foundPath := false
	for _, h := range headers.Headers.Headers {
		if h.Key == ":method" {
			assert.Equal(t, "GET", h.Value)
			foundMethod = true
		}
		if h.Key == ":path" {
			assert.Equal(t, "/api/test", h.Value)
			foundPath = true
		}
	}
	assert.True(t, foundMethod)
	assert.True(t, foundPath)
}

func TestBuildRequestHeaders_WithSchemeAndAuthority(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method:    "POST",
		Path:      "/api/test",
		Scheme:    "https",
		Authority: "example.com",
	}

	procReq := buildRequestHeaders(req)
	headers := procReq.GetRequestHeaders()
	require.NotNil(t, headers)
	require.NotNil(t, headers.Headers)

	foundScheme := false
	foundAuthority := false
	for _, h := range headers.Headers.Headers {
		if h.Key == ":scheme" {
			assert.Equal(t, "https", h.Value)
			foundScheme = true
		}
		if h.Key == ":authority" {
			assert.Equal(t, "example.com", h.Value)
			foundAuthority = true
		}
	}
	assert.True(t, foundScheme)
	assert.True(t, foundAuthority)
}

func TestBuildRequestHeaders_WithHeaders(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method: "GET",
		Path:   "/api/test",
		Headers: map[string]string{
			"content-type":  "application/json",
			"authorization": "Bearer token",
		},
	}

	procReq := buildRequestHeaders(req)
	headers := procReq.GetRequestHeaders()
	require.NotNil(t, headers)
	require.NotNil(t, headers.Headers)

	foundContentType := false
	foundAuth := false
	for _, h := range headers.Headers.Headers {
		if h.Key == "content-type" {
			assert.Equal(t, "application/json", h.Value)
			foundContentType = true
		}
		if h.Key == "authorization" {
			assert.Equal(t, "Bearer token", h.Value)
			foundAuth = true
		}
	}
	assert.True(t, foundContentType)
	assert.True(t, foundAuth)
}

func TestBuildRequestHeaders_ProcessRequestBody(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method:             "POST",
		Path:               "/api/test",
		ProcessRequestBody: true,
		Body:               []byte("test body"),
	}

	procReq := buildRequestHeaders(req)
	headers := procReq.GetRequestHeaders()
	require.NotNil(t, headers)
	assert.False(t, headers.EndOfStream)
}

func TestBuildRequestHeaders_ProcessRequestTrailers(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method:                 "POST",
		Path:                   "/api/test",
		ProcessRequestTrailers: true,
		Trailers: map[string]string{
			"x-checksum": "abc123",
		},
	}

	procReq := buildRequestHeaders(req)
	headers := procReq.GetRequestHeaders()
	require.NotNil(t, headers)
	assert.False(t, headers.EndOfStream)
}

func TestBuildRequestBody(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Body: []byte("test body content"),
	}

	procReq := buildRequestBody(req)
	assert.NotNil(t, procReq)

	body := procReq.GetRequestBody()
	assert.NotNil(t, body)
	assert.Equal(t, []byte("test body content"), body.Body)
	assert.True(t, body.EndOfStream)
}

func TestBuildRequestBody_WithTrailers(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Body:                   []byte("test body"),
		ProcessRequestTrailers: true,
	}

	procReq := buildRequestBody(req)
	body := procReq.GetRequestBody()
	require.NotNil(t, body)
	assert.False(t, body.EndOfStream)
}

func TestBuildRequestTrailers(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Trailers: map[string]string{
			"x-checksum":  "abc123",
			"x-signature": "xyz789",
		},
	}

	procReq := buildRequestTrailers(req)
	assert.NotNil(t, procReq)

	trailers := procReq.GetRequestTrailers()
	assert.NotNil(t, trailers)
	assert.NotNil(t, trailers.Trailers)
	assert.Len(t, trailers.Trailers.Headers, 2)
}

func TestBuildRequestTrailers_Empty(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Trailers: map[string]string{},
	}

	procReq := buildRequestTrailers(req)
	trailers := procReq.GetRequestTrailers()
	assert.NotNil(t, trailers)
	assert.Empty(t, trailers.Trailers.Headers)
}

func TestProcessingResult_Types(t *testing.T) {
	result := &ProcessingResult{
		Responses: []*PhaseResponse{
			{
				Phase: extproctorv1.ProcessingPhase_REQUEST_HEADERS,
				Response: &extprocv3.ProcessingResponse{
					Response: &extprocv3.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extprocv3.HeadersResponse{},
					},
				},
			},
		},
	}

	assert.Len(t, result.Responses, 1)
	assert.Equal(t, extproctorv1.ProcessingPhase_REQUEST_HEADERS, result.Responses[0].Phase)
}

func TestPhaseResponse_Types(t *testing.T) {
	resp := &PhaseResponse{
		Phase: extproctorv1.ProcessingPhase_REQUEST_BODY,
		Response: &extprocv3.ProcessingResponse{
			Response: &extprocv3.ProcessingResponse_RequestBody{
				RequestBody: &extprocv3.BodyResponse{},
			},
		},
	}

	assert.Equal(t, extproctorv1.ProcessingPhase_REQUEST_BODY, resp.Phase)
	assert.NotNil(t, resp.Response.GetRequestBody())
}

func TestClientConfig_Defaults(t *testing.T) {
	cfg := &clientConfig{
		target: "localhost:50051",
	}
	assert.Equal(t, "localhost:50051", cfg.target)
	assert.Empty(t, cfg.unixSocket)
	assert.False(t, cfg.tls)
}

func TestNew_DefaultTarget(t *testing.T) {
	client, err := New()
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Equal(t, "localhost:50051", client.Target())
}

func TestNew_WithTarget(t *testing.T) {
	client, err := New(WithTarget("localhost:9999"))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Equal(t, "localhost:9999", client.Target())
}

func TestNew_WithUnixSocket(t *testing.T) {
	client, err := New(WithUnixSocket("/tmp/test.sock"))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Equal(t, "unix:///tmp/test.sock", client.Target())
}

func TestClient_Close_WithConn(t *testing.T) {
	client, err := New()
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestBuildTLSConfig_NoCA(t *testing.T) {
	cfg := &clientConfig{
		tls: true,
	}

	tlsConfig, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, tlsConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
}

func TestBuildTLSConfig_InvalidCertPath(t *testing.T) {
	cfg := &clientConfig{
		tls:     true,
		tlsCert: "/nonexistent/cert.pem",
		tlsKey:  "/nonexistent/key.pem",
	}

	_, err := buildTLSConfig(cfg)
	assert.Error(t, err)
}

func TestBuildTLSConfig_InvalidCAPath(t *testing.T) {
	cfg := &clientConfig{
		tls:   true,
		tlsCA: "/nonexistent/ca.pem",
	}

	_, err := buildTLSConfig(cfg)
	assert.Error(t, err)
}

func TestBuildTLSConfig_InvalidCAPEM(t *testing.T) {
	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "invalid-ca.pem")
	err := os.WriteFile(caPath, []byte("not a valid certificate"), 0o644)
	require.NoError(t, err)

	cfg := &clientConfig{
		tls:   true,
		tlsCA: caPath,
	}

	_, err = buildTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CA certificate")
}

func TestNew_WithTLS_NoCerts(t *testing.T) {
	// Test TLS with no certificates (will use system certs)
	client, err := New(
		WithTarget("localhost:9999"),
		WithTLS("", "", ""),
	)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()
	assert.NotNil(t, client)
}

func TestNew_WithTLS_InvalidCert(t *testing.T) {
	_, err := New(
		WithTarget("localhost:9999"),
		WithTLS("/nonexistent/cert.pem", "/nonexistent/key.pem", ""),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TLS")
}

func TestNew_WithTLS_InvalidCA(t *testing.T) {
	_, err := New(
		WithTarget("localhost:9999"),
		WithTLS("", "", "/nonexistent/ca.pem"),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TLS")
}

func TestNew_WithMultipleOptions(t *testing.T) {
	client, err := New(
		WithTarget("localhost:8080"),
	)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Equal(t, "localhost:8080", client.Target())
}

func TestBuildRequestHeaders_AllOptions(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Method:                 "POST",
		Path:                   "/api/data",
		Scheme:                 "https",
		Authority:              "api.example.com",
		ProcessRequestBody:     true,
		ProcessRequestTrailers: true,
		Headers: map[string]string{
			"content-type": "application/json",
			"x-api-key":    "secret",
		},
		Body: []byte("{}"),
	}

	procReq := buildRequestHeaders(req)
	headers := procReq.GetRequestHeaders()
	require.NotNil(t, headers)
	require.NotNil(t, headers.Headers)

	assert.False(t, headers.EndOfStream)
	assert.NotEmpty(t, headers.Headers.Headers)
}

func TestBuildRequestBody_AllOptions(t *testing.T) {
	req := &extproctorv1.HttpRequest{
		Body:                   []byte("request body content"),
		ProcessRequestTrailers: false,
	}

	procReq := buildRequestBody(req)
	body := procReq.GetRequestBody()
	require.NotNil(t, body)

	assert.True(t, body.EndOfStream)
	assert.Equal(t, []byte("request body content"), body.Body)
}

func TestBuildTLSConfig_WithValidCerts(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate a self-signed certificate for testing
	certPEM, keyPEM := generateTestCertificate(t)

	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	err := os.WriteFile(certPath, certPEM, 0o644)
	require.NoError(t, err)
	err = os.WriteFile(keyPath, keyPEM, 0o644)
	require.NoError(t, err)

	cfg := &clientConfig{
		tls:     true,
		tlsCert: certPath,
		tlsKey:  keyPath,
	}

	tlsConfig, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, tlsConfig)
	assert.Len(t, tlsConfig.Certificates, 1)
}

func TestBuildTLSConfig_WithValidCA(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate a CA certificate for testing
	certPEM, _ := generateTestCertificate(t)

	caPath := filepath.Join(tmpDir, "ca.pem")
	err := os.WriteFile(caPath, certPEM, 0o644)
	require.NoError(t, err)

	cfg := &clientConfig{
		tls:   true,
		tlsCA: caPath,
	}

	tlsConfig, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, tlsConfig)
	assert.NotNil(t, tlsConfig.RootCAs)
}

// generateTestCertificate generates a self-signed certificate for testing
func generateTestCertificate(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	// Encode to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certPEM, keyPEM
}
