package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/JIeeiroSst/utils/logger"
)

type CloudflareProxy struct {
	ApiToken  string
	AccountId string
}

func NewCloudflareProxy(apiToken string,
	accountId string) *CloudflareProxy {
	return &CloudflareProxy{
		ApiToken:  apiToken,
		AccountId: accountId,
	}
}

const (
	cloudflareAPIURL    = "https://api.cloudflare.com/client/v4/accounts/%s/stream/direct_upload"
	authorizationHeader = "Bearer %s"
	maxDurationSeconds  = 3600
)

func (c *CloudflareProxy) GenerateUnique(ctx context.Context) (*GenerateUnique, error) {
	requestBody := directUploadRequest{
		MaxDurationSeconds: maxDurationSeconds,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error(ctx, "Error marshalling request body: %v", err)
		return nil, err
	}
	url := fmt.Sprintf(cloudflareAPIURL, c.AccountId)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBodyBytes))
	if err != nil {
		logger.Error(ctx, "Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf(authorizationHeader, c.ApiToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "Error reading response body: %v", err)
		return nil, err
	}

	var result GenerateUniqueResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		logger.Error(ctx, "Error reading unmarshal body: %v", err)
		return nil, err
	}

	if !result.Success {
		logger.Error(ctx, "Error call api: %v", result.Success)
		return nil, err
	}

	return &result.Result, nil
}

func (c *CloudflareProxy) UploadVideo(ctx context.Context, body *bytes.Buffer) (string, error) {
	generateUnique, err := c.GenerateUnique(ctx)
	if err != nil {
		logger.Error(ctx, "Error creating request: %v", err)
		return "", err
	}
	if generateUnique == nil {
		logger.Error(ctx, "Error generateUnique nil: %v", err)
		return "", err
	}
	writer := multipart.NewWriter(body)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://upload.videodelivery.net/%v", generateUnique.UID), body)
	if err != nil {
		logger.Error(ctx, "Error creating request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "Error sending request: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	return generateUnique.UploadURL, nil
}
