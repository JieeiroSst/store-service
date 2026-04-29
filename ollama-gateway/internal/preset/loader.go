package preset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type QA struct {
	ID       string   `json:"id"`
	Keywords []string `json:"keywords"` // for full-text search
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	ImageURL string `json:"image_url,omitempty"`
}

type PresetFile struct {
	Category string `json:"category"`
	Items    []QA   `json:"items"`
}

type Loader struct {
	dir    string
	presets []QA
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

func (l *Loader) Load() ([]QA, error) {
	if _, err := os.Stat(l.dir); os.IsNotExist(err) {
		if err := os.MkdirAll(l.dir, 0755); err != nil {
			return nil, fmt.Errorf("create presets dir: %w", err)
		}
		if err := l.createSamplePreset(); err != nil {
			return nil, err
		}
	}

	var all []QA
	err := filepath.WalkDir(l.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		items, err := l.loadFile(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		all = append(all, items...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.presets = all
	return all, nil
}

func (l *Loader) loadFile(path string) ([]QA, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pf PresetFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}

	for i := range pf.Items {
		if pf.Items[i].ID == "" {
			pf.Items[i].ID = fmt.Sprintf("%s_%d", pf.Category, i)
		}
	}
	return pf.Items, nil
}

func (l *Loader) createSamplePreset() error {
	sample := PresetFile{
		Category: "general",
		Items: []QA{
			{
				ID:       "greeting_1",
				Keywords: []string{"hello", "hi", "xin chào", "chào"},
				Question: "Xin chào / Hello",
				Answer:   "Xin chào! Tôi là AI Assistant. Tôi có thể giúp gì cho bạn?",
			},
			{
				ID:       "hours_1",
				Keywords: []string{"giờ làm việc", "mở cửa", "working hours", "open"},
				Question: "Giờ làm việc?",
				Answer:   "Chúng tôi làm việc từ 8:00 - 17:30, Thứ 2 đến Thứ 6.",
			},
			{
				ID:       "contact_1",
				Keywords: []string{"liên hệ", "contact", "phone", "email", "số điện thoại"},
				Question: "Thông tin liên hệ",
				Answer:   "Email: support@company.com\nHotline: 1900-xxxx",
			},
		},
	}

	data, _ := json.MarshalIndent(sample, "", "  ")
	return os.WriteFile(filepath.Join(l.dir, "general.json"), data, 0644)
}
