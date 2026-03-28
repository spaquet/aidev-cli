package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// SSEClient manages a Server-Sent Events connection
type SSEClient struct {
	client    *http.Client
	baseURL   string
	token     string
	ctx       context.Context
	cancel    context.CancelFunc
	eventCh   chan SSEEvent
	closeCh   chan struct{}
	closed    bool
}

// NewSSEClient creates a new SSE client
func NewSSEClient(baseURL, token string) *SSEClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &SSEClient{
		client:  &http.Client{Timeout: 0}, // No timeout for streaming
		baseURL: baseURL,
		token:   token,
		ctx:     ctx,
		cancel:  cancel,
		eventCh: make(chan SSEEvent, 10),
		closeCh: make(chan struct{}),
	}
}

// Start begins listening for SSE events
func (s *SSEClient) Start() error {
	go s.listen()
	return nil
}

// Events returns the event channel
func (s *SSEClient) Events() <-chan SSEEvent {
	return s.eventCh
}

// Close closes the SSE connection
func (s *SSEClient) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true
	s.cancel()
	close(s.closeCh)
	return nil
}

// Private methods

func (s *SSEClient) listen() {
	defer close(s.eventCh)

	url := s.baseURL + "/api/v1/instances/events"

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		err := s.connect(url)
		if err != nil && !s.closed {
			// Reconnect with backoff
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		if s.closed {
			return
		}
	}
}

func (s *SSEClient) connect(url string) error {
	req, err := http.NewRequestWithContext(s.ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE connection failed: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse event
		if strings.HasPrefix(line, "event: ") {
			eventType := strings.TrimPrefix(line, "event: ")

			// Read data line
			if scanner.Scan() {
				dataLine := scanner.Text()
				if strings.HasPrefix(dataLine, "data: ") {
					jsonData := strings.TrimPrefix(dataLine, "data: ")

					var data map[string]interface{}
					if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
						continue
					}

					event := SSEEvent{
						Type: eventType,
						Data: data,
					}

					select {
					case s.eventCh <- event:
					case <-s.ctx.Done():
						return nil
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
