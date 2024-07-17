package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gometheus/pkg/request"
	"github.com/m1khal3v/gometheus/pkg/response"
	"github.com/m1khal3v/gometheus/pkg/retry"
	"hash"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// This is a very simplified regular expression that will work in most cases.
// In border cases, you can disable address verification through the config
var addressRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9][a-zA-Z0-9-.]*(:\d+)?(/[a-zA-Z0-9-_+%]*)*$`)

type Client struct {
	resty *resty.Client
	retry bool
}

type SignatureConfig struct {
	Key    string
	Hash   func() hash.Hash
	Header string
}

type Config struct {
	Address   string
	Transport http.RoundTripper
	Signature *SignatureConfig

	DisableCompress          bool
	DisableAddressValidation bool
	DisableRetry             bool
}

type preRequestHook func(config *Config, request *http.Request) error

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.Status)
}

func newErrUnexpectedStatus(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

type ErrInvalidAddress struct {
	Address string
}

func (err ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid address: %s", err.Address)
}

func newErrInvalidAddress(address string) ErrInvalidAddress {
	return ErrInvalidAddress{
		Address: address,
	}
}

func New(config *Config) (*Client, error) {
	if err := prepareConfig(config); err != nil {
		return nil, err
	}

	client := resty.
		New().
		SetTransport(config.Transport).
		SetBaseURL(config.Address).
		SetHeader("Accept-Encoding", "gzip")

	hooks := make([]preRequestHook, 0)
	if !config.DisableCompress {
		hooks = append(hooks, compressRequestBody)
	}
	if config.Signature.Key != "" && config.Signature.Hash != nil && config.Signature.Header != "" {
		hooks = append(hooks, addHMACSignature)
	}

	client.SetPreRequestHook(preRequestHookCombine(config, hooks...))

	return &Client{resty: client, retry: !config.DisableRetry}, nil
}

func prepareConfig(config *Config) error {
	if !config.DisableAddressValidation {
		if !strings.HasPrefix(config.Address, "http") {
			config.Address = "http://" + config.Address
		}

		if !addressRegex.MatchString(config.Address) {
			return newErrInvalidAddress(config.Address)
		}
	}

	if config.Transport == nil {
		config.Transport = http.DefaultTransport
	}

	return nil
}

func preRequestHookCombine(config *Config, functions ...preRequestHook) resty.PreRequestHook {
	return func(client *resty.Client, request *http.Request) error {
		for _, function := range functions {
			if err := function(config, request); err != nil {
				return err
			}
		}

		return nil
	}
}

func compressRequestBody(config *Config, request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	writer, err := gzip.NewWriterLevel(buffer, 5)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, request.Body)
	if err = errors.Join(err, writer.Close(), request.Body.Close()); err != nil {
		return err
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return nil
}

func addHMACSignature(config *Config, request *http.Request) error {
	buffer := bytes.NewBuffer([]byte{})

	if request.Body != nil {
		_, err := io.Copy(buffer, request.Body)
		if err = errors.Join(err, request.Body.Close()); err != nil {
			return err
		}
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}

	request.Header.Set(config.Signature.Header, hex.EncodeToString(hmac.
		New(config.Signature.Hash, []byte(config.Signature.Key)).
		Sum(buffer.Bytes()),
	))

	return nil
}

func (client *Client) SaveMetric(ctx context.Context, metricType, metricName, metricValue string) (*response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"type":  metricType,
			"name":  metricName,
			"value": metricValue,
		}).
		SetError(&response.APIError{}),
		resty.MethodPost, "update/{type}/{name}/{value}")

	if err != nil {
		if result.RawResponse == nil {
			return nil, err
		} else {
			return result.Error().(*response.APIError), err
		}
	}

	return nil, nil
}

func (client *Client) SaveMetricAsJSON(ctx context.Context, request *request.SaveMetricRequest) (*response.SaveMetricResponse, *response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(&response.SaveMetricResponse{}).
		SetError(&response.APIError{}),
		resty.MethodPost, "update")

	if err != nil {
		if result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return result.Result().(*response.SaveMetricResponse), nil, nil
}

func (client *Client) SaveMetricsAsJSON(ctx context.Context, requests []request.SaveMetricRequest) ([]response.SaveMetricResponse, *response.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(requests).
		SetResult(&[]response.SaveMetricResponse{}).
		SetError(&response.APIError{}),
		resty.MethodPost, "updates")

	if err != nil {
		if result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*response.APIError), err
		}
	}

	return *result.Result().(*[]response.SaveMetricResponse), nil, nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx)
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response = nil
	do := func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		if result.StatusCode() != http.StatusOK {
			return newErrUnexpectedStatus(result.StatusCode())
		}

		return nil
	}

	var err error
	if !client.retry {
		err = do()
	} else {
		err = retry.Retry(time.Second, 5*time.Second, 4, 2, do, func(err error) bool {
			return !errors.As(err, &ErrUnexpectedStatus{})
		})
	}

	return result, err
}
