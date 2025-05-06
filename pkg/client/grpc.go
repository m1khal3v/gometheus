package client

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"hash"
	"net"
	"sync"

	"github.com/m1khal3v/gometheus/pkg/proto"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GRPCClient struct {
	conn     *grpc.ClientConn
	client   proto.MetricsServiceClient
	hmacPool *sync.Pool
	config   *config
	realIP   net.IP
}

func NewGRPC(address string, options ...ConfigOption) (*GRPCClient, error) {
	cfg := newConfig(address, options...)

	grpcOpts := make([]grpc.DialOption, 0, 1)

	if cfg.publicKey == nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				if len(rawCerts) == 0 {
					return errors.New("server certificate not provided")
				}

				cert, err := x509.ParseCertificate(rawCerts[0])
				if err != nil {
					return err
				}

				serverPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
				if !ok {
					return errors.New("server public key is not RSA")
				}

				if !serverPubKey.Equal(cfg.publicKey) {
					return errors.New("server public key does not match configured key")
				}
				return nil
			},
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	if cfg.compress {
		grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	}

	conn, err := grpc.NewClient(cfg.baseURL.Host, grpcOpts...)
	if err != nil {
		return nil, err
	}

	client := &GRPCClient{
		conn:   conn,
		client: proto.NewMetricsServiceClient(conn),
		config: cfg,
	}

	if cfg.signature != nil {
		client.hmacPool = &sync.Pool{
			New: func() any {
				return hmac.New(cfg.signature.hasher, []byte(cfg.signature.key))
			},
		}
	}

	return client, nil
}

func (c *GRPCClient) SaveMetric(ctx context.Context, req *request.SaveMetricRequest) (*response.SaveMetricResponse, *response.APIError, error) {
	grpcReq := c.convertRequest(req)

	ctx = c.addHeaders(ctx, grpcReq)
	resp, err := c.client.SaveMetric(ctx, grpcReq)
	if err != nil {
		return nil, convertError(err), err
	}
	return c.convertResponse(resp), nil, nil
}

func (c *GRPCClient) SaveMetrics(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error) {
	batch := make([]*proto.SaveMetricRequest, 0, len(requests))
	for _, req := range requests {
		batch = append(batch, c.convertRequest(&req))
	}

	req := &proto.SaveMetricsBatchRequest{Metrics: batch}
	ctx = c.addHeaders(ctx, req)
	resp, err := c.client.SaveMetrics(ctx, req)
	if err != nil {
		return nil, convertError(err), err
	}

	results := make([]response.SaveMetricResponse, 0, len(resp.Metrics))
	for _, m := range resp.Metrics {
		results = append(results, *c.convertResponse(m))
	}
	return results, nil, nil
}

func (c *GRPCClient) addHeaders(ctx context.Context, req gproto.Message) context.Context {
	md := metadata.New(nil)

	if c.config.signature != nil {
		encoder := c.hmacPool.Get().(hash.Hash)
		defer c.hmacPool.Put(encoder)
		encoder.Reset()

		data, _ := gproto.Marshal(req)
		encoder.Write(data)
		signature := hex.EncodeToString(encoder.Sum(nil))

		md.Set(c.config.signature.header, signature)
	}

	realIP, _ := c.getRealIP()
	md.Set("X-Real-IP", realIP.String())

	return metadata.NewOutgoingContext(ctx, md)
}

func (c *GRPCClient) convertRequest(req *request.SaveMetricRequest) *proto.SaveMetricRequest {
	request := &proto.SaveMetricRequest{
		MetricName: req.MetricName,
		MetricType: req.MetricType,
	}

	if nil != req.Delta {
		request.Delta = wrapperspb.Int64(*req.Delta)
	}

	if nil != req.Value {
		request.Value = wrapperspb.Double(*req.Value)
	}

	return request
}

func (c *GRPCClient) convertResponse(resp *proto.SaveMetricResponse) *response.SaveMetricResponse {
	response := &response.SaveMetricResponse{
		MetricName: resp.MetricName,
		MetricType: resp.MetricType,
	}

	if nil != resp.Delta {
		response.Delta = &resp.Delta.Value
	}

	if nil != resp.Value {
		response.Value = &resp.Value.Value
	}

	return response
}

func (c *GRPCClient) getRealIP() (net.IP, error) {
	if c.realIP != nil {
		return c.realIP, nil
	}

	port := "80"
	if c.config.baseURL.Port() != "" {
		port = c.config.baseURL.Port()
	}

	conn, err := net.Dial("udp", c.config.baseURL.Hostname()+":"+port)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c.realIP = conn.LocalAddr().(*net.UDPAddr).IP

	return c.realIP, nil
}

func convertError(err error) *response.APIError {
	return &response.APIError{Code: 500, Message: err.Error()}
}
