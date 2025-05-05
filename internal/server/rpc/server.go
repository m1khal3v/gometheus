package rpc

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"hash"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
)

type serverConfig struct {
	hmacSecret      string
	signatureHeader string
	hasher          func() hash.Hash
	privateKey      *rsa.PrivateKey
	allowedSubnet   *net.IPNet
}

type ServerOption func(*serverConfig)

func WithHMAC(secret, header string, hasher func() hash.Hash) ServerOption {
	return func(c *serverConfig) {
		c.hmacSecret = secret
		c.signatureHeader = header
		c.hasher = hasher
	}
}

func WithTLS(privateKey *rsa.PrivateKey) ServerOption {
	return func(c *serverConfig) {
		c.privateKey = privateKey
	}
}

func WithSubnet(header string, subnet *net.IPNet) ServerOption {
	return func(c *serverConfig) {
		c.allowedSubnet = subnet
	}
}

type GRPCServer struct {
	server   *grpc.Server
	config   *serverConfig
	hmacPool *sync.Pool
}

func NewGRPCServer(storage storage.Storage, options ...ServerOption) (*GRPCServer, error) {
	cfg := &serverConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	serverOpts := []grpc.ServerOption{}

	if cfg.privateKey != nil {
		cert, err := generateSelfSignedCert(cfg.privateKey)
		if err != nil {
			return nil, err
		}

		creds := credentials.NewServerTLSFromCert(cert)
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	if cfg.hmacSecret != "" {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(hmacInterceptor(cfg)))
	}

	if cfg.allowedSubnet != nil {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(subnetInterceptor("X-Real-IP", cfg.allowedSubnet)))
	}

	server := grpc.NewServer(serverOpts...)
	proto.RegisterMetricsServiceServer(server, NewMetricsService(storage))

	return &GRPCServer{
		server: server,
		config: cfg,
	}, nil
}

func hmacInterceptor(cfg *serverConfig) grpc.UnaryServerInterceptor {
	pool := &sync.Pool{
		New: func() interface{} {
			return hmac.New(cfg.hasher, []byte(cfg.hmacSecret))
		},
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		signatures := md.Get(cfg.signatureHeader)
		if len(signatures) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing signature")
		}

		raw, err := gproto.Marshal(req.(gproto.Message))
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal request")
		}

		h := pool.Get().(hash.Hash)
		defer pool.Put(h)
		h.Reset()

		if _, err := h.Write(raw); err != nil {
			return nil, status.Error(codes.Internal, "failed to compute signature")
		}

		expected := hex.EncodeToString(h.Sum(nil))
		if !hmac.Equal([]byte(signatures[0]), []byte(expected)) {
			return nil, status.Error(codes.Unauthenticated, "invalid signature")
		}

		return handler(ctx, req)
	}
}

func generateSelfSignedCert(privateKey *rsa.PrivateKey) (*tls.Certificate, error) {
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}, nil
}

func subnetInterceptor(header string, subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var ipStr string

		// Пытаемся получить IP из метаданных
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get(header)
			if len(values) > 0 {
				ipStr = values[0]
			}
		}

		// Если не нашли в метаданных, пробуем получить из peer
		if ipStr == "" {
			if p, ok := peer.FromContext(ctx); ok {
				ipStr = p.Addr.String()
			}
		}

		if ipStr == "" {
			return nil, status.Error(codes.PermissionDenied, "IP address not found")
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "Invalid IP format")
		}

		if !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "IP not in allowed subnet")
		}

		return handler(ctx, req)
	}
}

func (s *GRPCServer) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
