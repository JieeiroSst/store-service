package learning

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	KindTestCases = "testcases"
	KindCode      = "code_review"
	KindQCRun     = "qc_run"

	VerdictGood = "good"
	VerdictBad  = "bad"

	SourceHuman = "human"
	SourceAuto  = "auto"
)

type Experience struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Input     string          `json:"input"`
	Output    json.RawMessage `json:"output"`
	Model     string          `json:"model,omitempty"`
	Verdict   string          `json:"verdict,omitempty"`
	Source    string          `json:"source,omitempty"`
	Score     float64         `json:"score,omitempty"`
	Note      string          `json:"note,omitempty"`
	Corrected json.RawMessage `json:"corrected,omitempty"`
	Signals   map[string]any  `json:"signals,omitempty"`
	Embedding []float32       `json:"embedding,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

func (e *Experience) Target() json.RawMessage {
	if len(e.Corrected) > 0 {
		return e.Corrected
	}
	return e.Output
}

type Store struct {
	mu                sync.RWMutex
	path              string
	playbookPath      string
	items             []*Experience
	playbook          string
	feedbackSinceLast int
	history           []Snapshot
	historyPath       string
	activePath        string
}

type Snapshot struct {
	At      time.Time            `json:"at"`
	Model   string               `json:"model"`
	Overall float64              `json:"approval_rate"`
	Human   float64              `json:"human_rate"`
	Auto    float64              `json:"auto_rate"`
	Reviews int                  `json:"reviews"`
	ByModel map[string]ModelRate `json:"by_model"`
}

type ModelRate struct {
	Good int     `json:"good"`
	Bad  int     `json:"bad"`
	Rate float64 `json:"rate"`
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		path: filepath.Join(dir, "experiences.jsonl"), playbookPath: filepath.Join(dir, "playbook.txt"),
		historyPath: filepath.Join(dir, "history.jsonl"), activePath: filepath.Join(dir, "active_model.txt"),
	}
	if f, err := os.Open(s.historyPath); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			var sn Snapshot
			if json.Unmarshal(sc.Bytes(), &sn) == nil {
				s.history = append(s.history, sn)
			}
		}
	}
	if f, err := os.Open(s.path); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 32<<20)
		for sc.Scan() {
			var e Experience
			if json.Unmarshal(sc.Bytes(), &e) == nil {
				s.items = append(s.items, &e)
			}
		}
	}
	if b, err := os.ReadFile(s.playbookPath); err == nil {
		s.playbook = string(b)
	}
	return s, nil
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) Add(e *Experience) error {
	e.ID, e.CreatedAt = newID(), time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, e)
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(e)
}

func (s *Store) Feedback(id, verdict, note string, corrected json.RawMessage) error {
	if verdict != VerdictGood && verdict != VerdictBad {
		return errors.New(`verdict must be "good" or "bad"`)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.items {
		if e.ID == id {
			e.Verdict, e.Note, e.Source = verdict, note, SourceHuman
			if len(corrected) > 0 {
				e.Corrected = corrected
			}
			s.feedbackSinceLast++
			return s.rewriteLocked()
		}
	}
	return errors.New("experience not found")
}

func (s *Store) rewriteLocked() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, e := range s.items {
		if err := enc.Encode(e); err != nil {
			f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Similar(kind string, query []float32, k int) []*Experience {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var pool []*Experience
	for _, e := range s.items {
		if e.Kind == kind && e.Verdict == VerdictGood {
			pool = append(pool, e)
		}
	}
	if len(query) > 0 {
		sort.SliceStable(pool, func(i, j int) bool { return cosine(query, pool[i].Embedding) > cosine(query, pool[j].Embedding) })
	} else {
		sort.SliceStable(pool, func(i, j int) bool { return pool[i].CreatedAt.After(pool[j].CreatedAt) })
	}
	if len(pool) > k {
		pool = pool[:k]
	}
	return pool
}

func (s *Store) WithFeedback(n int) []*Experience {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Experience
	for i := len(s.items) - 1; i >= 0 && len(out) < n; i-- {
		if s.items[i].Verdict != "" {
			out = append(out, s.items[i])
		}
	}
	return out
}

func (s *Store) Playbook() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.playbook
}

func (s *Store) SetPlaybook(p string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playbook, s.feedbackSinceLast = p, 0
	return os.WriteFile(s.playbookPath, []byte(p), 0o644)
}

func (s *Store) FeedbackSinceReflect() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.feedbackSinceLast
}

func (s *Store) Approved() []*Experience {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Experience
	for _, e := range s.items {
		if e.Verdict == VerdictGood {
			out = append(out, e)
		}
	}
	return out
}

type Stats struct {
	Total                int     `json:"total"`
	Good                 int     `json:"good"`
	Bad                  int     `json:"bad"`
	Pending              int     `json:"pending"`
	ApprovalRate         float64 `json:"approval_rate"`
	FeedbackSinceReflect int     `json:"feedback_since_reflect"`
	Playbook             string  `json:"playbook"`
}

func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := Stats{Total: len(s.items), FeedbackSinceReflect: s.feedbackSinceLast, Playbook: s.playbook}
	for _, e := range s.items {
		switch e.Verdict {
		case VerdictGood:
			st.Good++
		case VerdictBad:
			st.Bad++
		default:
			st.Pending++
		}
	}
	if r := st.Good + st.Bad; r > 0 {
		st.ApprovalRate = math.Round(float64(st.Good)/float64(r)*1000) / 1000
	}
	return st
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return -1
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i] * b[i])
		na += float64(a[i] * a[i])
		nb += float64(b[i] * b[i])
	}
	if na == 0 || nb == 0 {
		return -1
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func (s *Store) AutoReview(id, verdict, note string, score float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.items {
		if e.ID != id {
			continue
		}
		if e.Source == SourceHuman {
			return nil
		}
		e.Score, e.Note = score, note
		if verdict != "" {
			e.Verdict, e.Source = verdict, SourceAuto
			s.feedbackSinceLast++
		}
		return s.rewriteLocked()
	}
	return errors.New("experience not found")
}

func (s *Store) Pending(n int) []*Experience {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Experience
	for _, e := range s.items {
		if e.Verdict == "" {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return math.Abs(out[i].Score-0.5) < math.Abs(out[j].Score-0.5)
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func rate(g, b int) float64 {
	if g+b == 0 {
		return 0
	}
	return math.Round(float64(g)/float64(g+b)*1000) / 1000
}

func (s *Store) RateByModel(humanOnly bool) map[string]ModelRate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rateLocked(humanOnly)
}

func (s *Store) rateLocked(humanOnly bool) map[string]ModelRate {
	out := map[string]ModelRate{}
	for _, e := range s.items {
		if e.Verdict == "" || (humanOnly && e.Source != SourceHuman) {
			continue
		}
		m := out[e.Model]
		if e.Verdict == VerdictGood {
			m.Good++
		} else {
			m.Bad++
		}
		out[e.Model] = m
	}
	for k, m := range out {
		m.Rate = rate(m.Good, m.Bad)
		out[k] = m
	}
	return out
}

func (s *Store) TakeSnapshot(activeModel string) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	var g, b, hg, hb, ag, ab int
	for _, e := range s.items {
		if e.Verdict == "" {
			continue
		}
		good := e.Verdict == VerdictGood
		switch {
		case good:
			g++
		default:
			b++
		}
		if e.Source == SourceHuman {
			if good {
				hg++
			} else {
				hb++
			}
		} else if good {
			ag++
		} else {
			ab++
		}
	}
	sn := Snapshot{At: time.Now().UTC(), Model: activeModel, Overall: rate(g, b), Human: rate(hg, hb), Auto: rate(ag, ab),
		Reviews: g + b, ByModel: s.rateLocked(false)}
	s.history = append(s.history, sn)
	if f, err := os.OpenFile(s.historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		_ = json.NewEncoder(f).Encode(sn)
		f.Close()
	}
	return sn
}

func (s *Store) History(n int) []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := s.history
	if len(h) > n {
		h = h[len(h)-n:]
	}
	return append([]Snapshot(nil), h...)
}

func (s *Store) ActiveModel() string {
	b, _ := os.ReadFile(s.activePath)
	return strings.TrimSpace(string(b))
}

func (s *Store) SetActiveModel(m string) error { return os.WriteFile(s.activePath, []byte(m), 0o644) }

func (s *Store) Get(id string) (*Experience, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.items {
		if e.ID == id {
			return e, true
		}
	}
	return nil, false
}
