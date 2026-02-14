package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	ip       string
	port     string
	listener net.Listener
}

func New(port string) (*Server, error) {
	addr := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("create listener on %s: %w", addr, err)
	}

	return &Server{
		ip:       listener.Addr().(*net.TCPAddr).IP.String(),
		port:     strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		listener: listener,
	}, nil
}

func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		slog.Debug("server.Serve: context closed")
		shuwdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		slog.Debug("server.Serve: shutting down")
		errCh <- srv.Shutdown(shuwdownCtx)
	}()

	if err := srv.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	slog.Debug("server.Serve: serving stopped")

	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func (s *Server) ServerHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           handler,
	})
}
