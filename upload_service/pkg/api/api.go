package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/JIeeiroSst/upload-service/common"
	"github.com/JIeeiroSst/upload-service/pkg/log"
)

type UploadApi struct {
	URL   string
	Token string
	Name  string
}

type UploadData struct {
	Status  string
	Success int
	Data
}

type Data struct {
	ID         string
	Title      string
	UrlViewer  string
	Url        string
	DisplayUrl string
	Width      string
	Height     string
	Size       string
	Time       string
	Expiration string
	Thumb
	Medium
	DeleteUrl string
}

type Image struct {
	Filename  string
	Name      string
	Mime      string
	Extension string
	URL       string
}

type Thumb struct {
	Filename  string
	Name      string
	Mime      string
	Extension string
	URl       string
}

type Medium struct {
	Filename  string
	Name      string
	Mime      string
	Extension string
	URl       string
}

func NewUploadFile(u *UploadApi) *UploadApi {
	return &UploadApi{
		URL:   u.URL,
		Token: u.Token,
	}
}

func (u *UploadApi) UploadFile(f multipart.File, h *multipart.FileHeader) (*UploadData, error) {
	var data UploadData
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("image", h.Filename)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)

	if err != nil {
		return nil, err
	}

	part.Write(b)
	writer.Close()

	req, _ := http.NewRequest("POST", u.URL, buf)

	req.Header.Add("Content-Type", writer.FormDataContentType())
	params := req.URL.Query()
	params.Add("key", u.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ = io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, errors.New(string(b))
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if data.Success == common.Success {
		log.Error(common.ApiFailed.Error())
		return nil, common.ApiFailed
	}
	return &data, nil
}
