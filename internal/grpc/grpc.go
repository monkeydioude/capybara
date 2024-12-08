package grpc

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/monkeydioude/capybara/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CatchAllProxy intercepts all incoming gRPC requests and forwards them to the backend
type CatchAllProxy struct {
	BackendPort int32
	Credentials credentials.TransportCredentials
}

// Proxy implements the generic proxy functionality
func (p *CatchAllProxy) Proxy(srv any, stream grpc.ServerStream) error {
	// Extract the full method name (e.g., /ServiceName/MethodName)
	methodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return status.Errorf(codes.Unavailable, "Unable to get method name from stream")
	}
	log.Printf("Received request for method: %s", methodName)

	// Extract incoming metadata
	md, _ := metadata.FromIncomingContext(stream.Context())
	log.Printf("Incoming metadata: %v", md)

	// Backend server address
	backendAddress := fmt.Sprintf("localhost:%d", p.BackendPort) // Replace with dynamic resolution if needed

	// // TLS credentials for the backend connection
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, // Skip verification for testing; remove this in production
	})

	// if err != nil {
	// 	return status.Errorf(codes.Unavailable, "Unable to connect to backend: %v", err)
	// }
	// Dial the backend server using grpc.NewClient
	conn, err := grpc.NewClient(
		backendAddress,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Printf("[ERR ] %s: %v", errors.ErrUnableToConnectToBackend, err)
		return status.Errorf(codes.Unavailable, "%s: %v", errors.ErrUnableToConnectToBackend, err)
	}
	defer conn.Close()

	// Forward the request to the backend
	clientStream, err := grpc.NewClientStream(
		stream.Context(),
		&grpc.StreamDesc{
			ServerStreams: true,
			ClientStreams: true,
		},
		conn,
		methodName,
	)
	if err != nil {
		log.Printf("[ERR ] %s: %v", errors.ErrUnableToCreateClientStream, err)
		return status.Errorf(codes.Unavailable, "%a: %v", errors.ErrUnableToCreateClientStream, err)
	}

	// Proxy all requests and responses between the client and backend
	errCh := make(chan error, 2)
	go func() {
		forwardStream(stream, clientStream, errCh)
	}()
	go func() {
		forwardStream(clientStream, stream, errCh)
	}()
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

func IsGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc"
}
