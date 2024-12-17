package cloudflare

import "bytes"

type GenerateUnique struct {
	UploadURL string `json:"uploadURL"`
	UID       string `json:"uid"`
}

type GenerateUniqueResult struct {
	Result  GenerateUnique `json:"result"`
	Success bool           `json:"success"`
}

type UploadVideoRequest struct {
	UID   string        `json:"uid"`
	Video *bytes.Buffer `json:"video"`
}

type directUploadRequest struct {
	MaxDurationSeconds int `json:"maxDurationSeconds"`
}
