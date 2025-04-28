package rpc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	// Генерация приватного ключа RSA.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Генерация самоподписанного сертификата.
	cert, err := generateSelfSignedCert(privateKey)
	if err != nil {
		t.Fatalf("failed to generate self-signed certificate: %v", err)
	}

	if cert == nil {
		t.Fatal("cert should not be nil")
	}
}

func TestSubnetInterceptor_AllowedIP(t *testing.T) {
	// Задаем подсеть, в которой допустимы адреса.
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("failed to parse subnet: %v", err)
	}

	interceptor := subnetInterceptor("X-Real-IP", subnet)

	// Создаем контекст с метаданными, содержащими допустимый IP-адрес.
	md := metadata.Pairs("X-Real-IP", "192.168.1.42")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Пустой обработчик, так как для теста достаточно факта вызова.
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("interceptor failed: %v", err)
	}

	if resp != "success" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestSubnetInterceptor_DeniedIP(t *testing.T) {
	// Задаем подсеть, в которой допустимы только определенные адреса.
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("failed to parse subnet: %v", err)
	}

	interceptor := subnetInterceptor("X-Real-IP", subnet)

	// Создаем контекст с метаданными, содержащими недопустимый IP-адрес.
	md := metadata.Pairs("X-Real-IP", "10.0.0.5")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Пустой обработчик, так как для теста достаточно факта вызова.
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err == nil {
		t.Fatal("expected error but got none")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("unexpected error code: %v", st.Code())
	}
}
