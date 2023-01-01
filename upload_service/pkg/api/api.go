package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func (u *UploadApi) UploadFile(b bytes.Buffer, w *multipart.Writer) (*UploadData, error) {
	var data UploadData
	req, err := http.NewRequest("POST", u.URL, &b)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	params := req.URL.Query()
	params.Add("key", u.Token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if data.Success == common.Success {
		log.Error(common.ApiFailed.Error())
		return nil, common.ApiFailed
	}
	return &data, nil
}
