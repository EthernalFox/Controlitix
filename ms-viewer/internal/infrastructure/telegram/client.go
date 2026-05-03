package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	ParseModeMarkdownV2 = "MarkdownV2"
)

type Client struct {
	apiURL string
	token  string
	http   *http.Client
	logger *slog.Logger

	rateMutex sync.Mutex
	lastSent  time.Time
}

type sendMessageRequest struct {
	ChatID                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

type DryRunClient struct {
	logger *slog.Logger
}

func NewClient(
	apiURL string,
	token string,
	timeout time.Duration,
	logger *slog.Logger,
) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		apiURL: strings.TrimRight(strings.TrimSpace(apiURL), "/"),
		token:  strings.TrimSpace(token),
		http: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

func NewDryRunClient(logger *slog.Logger) *DryRunClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &DryRunClient{logger: logger}
}

func (client *Client) SendMessage(
	ctx context.Context,
	chatID int64,
	text string,
	parseMode string,
) error {
	if strings.TrimSpace(client.token) == "" {
		return &domain.PermanentSendError{Cause: fmt.Errorf("bot token is empty")}
	}

	if waitError := client.waitRateLimit(ctx); waitError != nil {
		return waitError
	}

	payload, marshalError := json.Marshal(sendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             parseMode,
		DisableWebPagePreview: true,
	})
	if marshalError != nil {
		return fmt.Errorf("marshal telegram request: %w", marshalError)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", client.apiURL, client.token)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if requestError != nil {
		return fmt.Errorf("build telegram request: %w", requestError)
	}
	request.Header.Set("Content-Type", "application/json")

	response, doError := client.http.Do(request)
	if doError != nil {
		return fmt.Errorf("send telegram request: %w", doError)
	}
	defer response.Body.Close()

	body, readError := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if readError != nil {
		return fmt.Errorf("read telegram response: %w", readError)
	}

	if response.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(response.Header.Get("Retry-After"))
		return &domain.RateLimitError{
			RetryAfter: retryAfter,
			Cause:      fmt.Errorf("telegram rate limited: status=%d", response.StatusCode),
		}
	}

	if response.StatusCode >= 400 && response.StatusCode < 500 {
		return &domain.PermanentSendError{Cause: fmt.Errorf(
			"telegram returned %d: %s",
			response.StatusCode,
			extractDescription(body),
		)}
	}

	if response.StatusCode >= 500 {
		return fmt.Errorf(
			"telegram returned %d: %s",
			response.StatusCode,
			extractDescription(body),
		)
	}

	var parsedResponse telegramResponse
	if unmarshalError := json.Unmarshal(body, &parsedResponse); unmarshalError == nil && !parsedResponse.OK {
		return &domain.PermanentSendError{Cause: fmt.Errorf(
			"telegram api rejected request: %s",
			strings.TrimSpace(parsedResponse.Description),
		)}
	}

	return nil
}

func (client *DryRunClient) SendMessage(
	_ context.Context,
	chatID int64,
	_ string,
	_ string,
) error {
	if client.logger != nil {
		client.logger.Info(
			"notifier dry run send",
			"method",
			"telegram.DryRunClient.SendMessage",
			"chat_id",
			chatID,
		)
	}
	return nil
}

func (client *Client) waitRateLimit(ctx context.Context) error {
	client.rateMutex.Lock()
	defer client.rateMutex.Unlock()

	now := time.Now().UTC()
	if client.lastSent.IsZero() {
		client.lastSent = now
		return nil
	}

	elapsed := now.Sub(client.lastSent)
	minInterval := 33 * time.Millisecond
	if elapsed >= minInterval {
		client.lastSent = now
		return nil
	}

	waitDuration := minInterval - elapsed
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		client.lastSent = time.Now().UTC()
		return nil
	}
}

func parseRetryAfter(value string) time.Duration {
	parsedValue, parseError := strconv.Atoi(strings.TrimSpace(value))
	if parseError != nil || parsedValue <= 0 {
		return 30 * time.Second
	}

	return time.Duration(parsedValue) * time.Second
}

func extractDescription(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var parsedResponse telegramResponse
	if unmarshalError := json.Unmarshal(body, &parsedResponse); unmarshalError == nil {
		if trimmed := strings.TrimSpace(parsedResponse.Description); trimmed != "" {
			return trimmed
		}
	}

	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) > 256 {
		return trimmed[:256]
	}
	return trimmed
}
