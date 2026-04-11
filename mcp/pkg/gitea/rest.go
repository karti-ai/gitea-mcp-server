package gitea

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mcpContext "gitea.com/gitea/gitea-mcp/pkg/context"
	"gitea.com/gitea/gitea-mcp/pkg/flag"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"go.uber.org/zap"
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("request failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("request failed with status %d: %s", e.StatusCode, e.Body)
}

func tokenFromContext(ctx context.Context) string {
	if ctx != nil {
		if token, ok := ctx.Value(mcpContext.TokenContextKey).(string); ok && token != "" {
			return token
		}
	}
	return flag.Token
}

func newRESTHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if flag.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested insecure mode
	}
	return &http.Client{
		Transport:     transport,
		Timeout:       60 * time.Second,
		CheckRedirect: checkRedirect,
	}
}

func buildAPIURL(path string, query url.Values) (string, error) {
	host := strings.TrimRight(flag.Host, "/")
	if host == "" {
		return "", errors.New("gitea host is empty")
	}
	p := strings.TrimLeft(path, "/")
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/%s", host, p))
	if err != nil {
		return "", err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

// DoJSON performs an API request and decodes a JSON response into respOut (if non-nil).
// It returns the HTTP status code.
func DoJSON(ctx context.Context, method, path string, query url.Values, body, respOut any) (int, error) {
	correlationID := log.GetCorrelationID(ctx)
	if correlationID == "" {
		ctx = log.WithCorrelationID(ctx, "")
		correlationID = log.GetCorrelationID(ctx)
	}

	operation := log.GetOperation(ctx)
	if operation == "" {
		operation = fmt.Sprintf("%s %s", method, path)
	}

	logger := log.WithContext(ctx)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			logger.Error("failed to marshal request body",
				zap.Error(err),
				zap.String("operation", operation),
				zap.String("method", method),
				zap.String("path", path),
			)
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	u, err := buildAPIURL(path, query)
	if err != nil {
		logger.Error("failed to build API URL",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("path", path),
		)
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		logger.Error("failed to create HTTP request",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("url", u),
		)
		return 0, fmt.Errorf("create request: %w", err)
	}

	token := tokenFromContext(ctx)
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := newRESTHTTPClient()

	logger.Debug("sending API request",
		zap.String("operation", operation),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("correlation_id", correlationID),
	)

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		logger.Error("API request failed",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Duration("duration", duration),
			zap.String("correlation_id", correlationID),
		)
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		logger.Error("API request returned error status",
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", resp.StatusCode),
			zap.Duration("duration", duration),
			zap.String("correlation_id", correlationID),
			zap.String("response_body", strings.TrimSpace(string(bodySnippet))),
		)
		return resp.StatusCode, &HTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(bodySnippet))}
	}

	logger.Debug("API request completed",
		zap.String("operation", operation),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration),
		zap.String("correlation_id", correlationID),
	)

	if respOut == nil {
		_, _ = io.Copy(io.Discard, resp.Body) // best-effort
		return resp.StatusCode, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(respOut); err != nil {
		logger.Error("failed to decode API response",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", resp.StatusCode),
		)
		return resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}
	return resp.StatusCode, nil
}

// DoBytes performs an API request and returns the raw response bytes.
// It returns the HTTP status code.
func DoBytes(ctx context.Context, method, path string, query url.Values, body any, accept string) ([]byte, int, error) {
	correlationID := log.GetCorrelationID(ctx)
	if correlationID == "" {
		ctx = log.WithCorrelationID(ctx, "")
		correlationID = log.GetCorrelationID(ctx)
	}

	operation := log.GetOperation(ctx)
	if operation == "" {
		operation = fmt.Sprintf("%s %s", method, path)
	}

	logger := log.WithContext(ctx)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			logger.Error("failed to marshal request body",
				zap.Error(err),
				zap.String("operation", operation),
				zap.String("method", method),
				zap.String("path", path),
			)
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	u, err := buildAPIURL(path, query)
	if err != nil {
		logger.Error("failed to build API URL",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("path", path),
		)
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		logger.Error("failed to create HTTP request",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("url", u),
		)
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	token := tokenFromContext(ctx)
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := newRESTHTTPClient()

	logger.Debug("sending API request",
		zap.String("operation", operation),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("correlation_id", correlationID),
	)

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		logger.Error("API request failed",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Duration("duration", duration),
			zap.String("correlation_id", correlationID),
		)
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodySnippet := respBytes
		if len(bodySnippet) > 8192 {
			bodySnippet = bodySnippet[:8192]
		}
		logger.Error("API request returned error status",
			zap.String("operation", operation),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", resp.StatusCode),
			zap.Duration("duration", duration),
			zap.String("correlation_id", correlationID),
			zap.String("response_body", strings.TrimSpace(string(bodySnippet))),
		)
		return nil, resp.StatusCode, &HTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(bodySnippet))}
	}

	logger.Debug("API request completed",
		zap.String("operation", operation),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration),
		zap.String("correlation_id", correlationID),
	)

	return respBytes, resp.StatusCode, nil
}
