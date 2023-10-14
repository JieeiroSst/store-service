package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func main() {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}
	req.Header.Add("Authorization", "Bearer <token>")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		return
	}

	fmt.Println(string(body))
}

var uploadPath string

type Media3Repository interface {
}

type ResponseMedia struct {}

type media3Repo struct {
	media3Host string
	client     *http.Client
}

func NewMedia3Repository(media3Host string) Media3Repository {
	return &media3Repo{
		media3Host: media3Host,
		client:     http.DefaultClient,
	}
}

func (m *media3Repo) fileUploadRequest(file []byte) (path string, err error) {
	url := fmt.Sprintf("%v%v", m.media3Host, uploadPath)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "file")
	if err != nil {
		return
	}
	io.Copy(fw, bytes.NewReader(file))

	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := m.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var media ResponseMedia

	err = json.Unmarshal(body.Bytes(), &media)
	if err != nil {
		return
	}
	// path = media.Data.URL
	return
}

// url := "http://localhost:3000/file/upload"
func (m *media3Repo) UploadImage(file string) (string, error) {
	res, err := http.Get(file)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("ioutil.ReadAll -> %v", err)
	}
	res.Body.Close()

	shareLink, err := m.fileUploadRequest(data)
	if err != nil {
		return "", nil
	}
	return shareLink, nil
}
