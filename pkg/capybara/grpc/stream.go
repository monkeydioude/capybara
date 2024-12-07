package grpc

import "io"

type streamOps interface {
	RecvMsg(m interface{}) error
	SendMsg(m interface{}) error
}

func forwardStream(src streamOps, dst streamOps, errCh chan error) {
	for {
		// Create a message for receiving
		m := &ProxyMessage{} // Ensure it matches your .proto file

		// Receive a message from the source stream
		err := src.RecvMsg(m)
		if err != nil {
			if err == io.EOF {
				// Forward EOF to the other stream
				errCh <- nil
				return
			}
			errCh <- err
			return
		}

		// Send the message to the destination stream
		err = dst.SendMsg(m)
		if err != nil {
			errCh <- err
			return
		}
	}
}
