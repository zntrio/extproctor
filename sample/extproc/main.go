// Package main implements a simple Envoy External Processor (ExtProc) filter.
// This gRPC service processes HTTP requests and responses flowing through Envoy.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// ExtProcServer implements the Envoy ExternalProcessor service.
type ExtProcServer struct {
	extprocv3.UnimplementedExternalProcessorServer
}

// Process handles the bidirectional streaming RPC for external processing.
func (s *ExtProcServer) Process(stream extprocv3.ExternalProcessor_ProcessServer) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive request: %v", err)
		}

		resp, err := s.processRequest(req)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to process request: %v", err)
		}

		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}
}

// processRequest routes the incoming request to the appropriate handler.
func (s *ExtProcServer) processRequest(req *extprocv3.ProcessingRequest) (*extprocv3.ProcessingResponse, error) {
	switch v := req.Request.(type) {
	case *extprocv3.ProcessingRequest_RequestHeaders:
		return s.handleRequestHeaders(v.RequestHeaders)
	case *extprocv3.ProcessingRequest_RequestBody:
		return s.handleRequestBody(v.RequestBody)
	case *extprocv3.ProcessingRequest_ResponseHeaders:
		return s.handleResponseHeaders(v.ResponseHeaders)
	case *extprocv3.ProcessingRequest_ResponseBody:
		return s.handleResponseBody(v.ResponseBody)
	default:
		// For unhandled message types, continue processing without modifications
		return &extprocv3.ProcessingResponse{}, nil
	}
}

// handleRequestHeaders processes incoming request headers.
// This is where you can inspect, modify, or add headers to the request.
func (s *ExtProcServer) handleRequestHeaders(headers *extprocv3.HttpHeaders) (*extprocv3.ProcessingResponse, error) {
	log.Printf("Processing request headers: method=%s path=%s",
		getHeader(headers, ":method"),
		getHeader(headers, ":path"))

	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{
				Response: &extprocv3.CommonResponse{
					// Continue processing the request
					Status: extprocv3.CommonResponse_CONTINUE,
					// Example: Add a custom header to the request
					HeaderMutation: &extprocv3.HeaderMutation{
						SetHeaders: []*corev3.HeaderValueOption{
							{
								Header: &corev3.HeaderValue{
									Key:   "x-extproc-processed",
									Value: "true",
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// handleRequestBody processes the request body.
// This is called when body processing is enabled in Envoy configuration.
func (s *ExtProcServer) handleRequestBody(body *extprocv3.HttpBody) (*extprocv3.ProcessingResponse, error) {
	log.Printf("Processing request body: size=%d end_of_stream=%v",
		len(body.Body), body.EndOfStream)

	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestBody{
			RequestBody: &extprocv3.BodyResponse{
				Response: &extprocv3.CommonResponse{
					Status: extprocv3.CommonResponse_CONTINUE,
				},
			},
		},
	}, nil
}

// handleResponseHeaders processes outgoing response headers.
// This is where you can inspect, modify, or add headers to the response.
func (s *ExtProcServer) handleResponseHeaders(headers *extprocv3.HttpHeaders) (*extprocv3.ProcessingResponse, error) {
	log.Printf("Processing response headers: status=%s",
		getHeader(headers, ":status"))

	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extprocv3.HeadersResponse{
				Response: &extprocv3.CommonResponse{
					Status: extprocv3.CommonResponse_CONTINUE,
					// Example: Add a custom header to the response
					HeaderMutation: &extprocv3.HeaderMutation{
						SetHeaders: []*corev3.HeaderValueOption{
							{
								Header: &corev3.HeaderValue{
									Key:   "x-extproc-response",
									Value: "processed",
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// handleResponseBody processes the response body.
// This is called when body processing is enabled in Envoy configuration.
func (s *ExtProcServer) handleResponseBody(body *extprocv3.HttpBody) (*extprocv3.ProcessingResponse, error) {
	log.Printf("Processing response body: size=%d end_of_stream=%v",
		len(body.Body), body.EndOfStream)

	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseBody{
			ResponseBody: &extprocv3.BodyResponse{
				Response: &extprocv3.CommonResponse{
					Status: extprocv3.CommonResponse_CONTINUE,
				},
			},
		},
	}, nil
}

// getHeader extracts a header value by key from the HttpHeaders message.
func getHeader(headers *extprocv3.HttpHeaders, key string) string {
	if headers == nil || headers.Headers == nil {
		return ""
	}
	for _, h := range headers.Headers.Headers {
		if h.Key == key {
			return string(h.RawValue)
		}
	}
	return ""
}

func main() {
	addr := flag.String("addr", ":50051", "gRPC server address")
	flag.Parse()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register ExtProc service
	extprocv3.RegisterExternalProcessorServer(grpcServer, &ExtProcServer{})

	// Register health service for load balancer health checks
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Create listener
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *addr, err)
	}

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	fmt.Printf("ExtProc server listening on %s\n", *addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
