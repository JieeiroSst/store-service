package queuecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail,omitempty"`
}

type RabbitMQ struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Options struct {
	MaxReady   int
	MaxUnacked int
}

type queue struct {
	Name          string `json:"name"`
	VHost         string `json:"vhost"`
	State         string `json:"state"`
	Messages      int    `json:"messages"`
	MessagesReady int    `json:"messages_ready"`
	MessagesUnack int    `json:"messages_unacknowledged"`
	Consumers     int    `json:"consumers"`
	Durable       bool   `json:"durable"`
	Arguments     map[string]any
	MessageStats  struct{ PublishDetails, DeliverDetails struct{ Rate float64 } } `json:"message_stats"`
}

var dlqName = regexp.MustCompile(`(?i)(dlq|dead|\.dl$|_dl$|error|failed|parking)`)

func CheckRabbitMQ(ctx context.Context, c *http.Client, cfg RabbitMQ, opt Options) ([]Finding, error) {
	if opt.MaxReady <= 0 {
		opt.MaxReady = 1000
	}
	if opt.MaxUnacked <= 0 {
		opt.MaxUnacked = 1000
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(cfg.URL, "/")+"/api/queues", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.User, cfg.Password)
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach management API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("management API rejected the credentials")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("management API returned %d", resp.StatusCode)
	}
	var qs []queue
	if err := json.NewDecoder(resp.Body).Decode(&qs); err != nil {
		return nil, fmt.Errorf("unexpected management API response: %w", err)
	}

	var out []Finding
	add := func(id, sev, title, detail string) { out = append(out, Finding{id, sev, title, detail}) }
	for _, q := range qs {
		name := q.VHost + q.Name
		switch {
		case dlqName.MatchString(q.Name) && q.Messages > 0:
			add("Q-DLQ-001", "high", fmt.Sprintf("dead-letter queue %q holds %d message(s)", q.Name, q.Messages), "Messages failed processing and were parked. Inspect and replay or fix the consumer.")
		case q.Consumers == 0 && q.Messages > 0:
			add("Q-CONS-001", "high", fmt.Sprintf("queue %q has %d message(s) and no consumers", name, q.Messages), "Nothing is processing this queue.")
		case q.MessagesReady > opt.MaxReady:
			add("Q-BACKLOG-001", "medium", fmt.Sprintf("queue %q backlog: %d ready messages", name, q.MessagesReady), fmt.Sprintf("Threshold %d; consumers: %d.", opt.MaxReady, q.Consumers))
		}
		if q.MessagesUnack > opt.MaxUnacked {
			add("Q-UNACK-001", "medium", fmt.Sprintf("queue %q has %d unacknowledged messages", name, q.MessagesUnack), "Consumers may be stuck or not acking.")
		}
		if !q.Durable && !strings.HasPrefix(q.Name, "amq.") && q.Messages > 0 {
			add("Q-DUR-001", "low", fmt.Sprintf("queue %q is not durable", name), "Messages are lost if the broker restarts.")
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return rank(out[i].Severity) < rank(out[j].Severity) })
	return out, nil
}

func rank(s string) int {
	switch s {
	case "high":
		return 0
	case "medium":
		return 1
	}
	return 2
}
