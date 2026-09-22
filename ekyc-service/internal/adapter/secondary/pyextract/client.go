package pyextract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Client struct {
	pythonBin  string
	scriptPath string
	timeout    time.Duration
}

func NewClient(pythonBin, scriptPath string, timeout time.Duration) *Client {
	return &Client{pythonBin: pythonBin, scriptPath: scriptPath, timeout: timeout}
}

func (c *Client) call(ctx context.Context, request map[string]any) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("pyextract: marshal request: %w", err)
	}

	cmd := exec.CommandContext(ctx, c.pythonBin, c.scriptPath)
	cmd.Stdin = bytes.NewReader(reqBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	var resp map[string]any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &resp); jsonErr != nil {
		if runErr != nil {
			return nil, fmt.Errorf("pyextract: %s: %s", runErr, stderr.String())
		}
		return nil, fmt.Errorf("pyextract: invalid JSON response: %w (stdout=%q stderr=%q)", jsonErr, stdout.String(), stderr.String())
	}
	if errMsg, ok := resp["error"].(string); ok {
		return nil, fmt.Errorf("pyextract: %s", errMsg)
	}
	return resp, nil
}
