package suite

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

type Approved struct {
	Hash       string           `json:"hash"`
	Case       apitest.TestCase `json:"case"`
	ApprovedAt time.Time        `json:"approved_at"`
	FromAI     bool             `json:"from_ai"`
}

type Suite struct {
	Name     string     `json:"name"`
	Cases    []Approved `json:"cases"`
	AIAccept int        `json:"ai_accepted"`
	AIReject int        `json:"ai_rejected"`
}

type Store struct {
	mu  sync.Mutex
	dir string
}

func Open(dir string) (*Store, error) {
	d := filepath.Join(dir, "suites")
	return &Store{dir: d}, os.MkdirAll(d, 0o755)
}

func ValidName(n string) bool { return nameRe.MatchString(n) }

func Hash(c apitest.TestCase) string {
	key := struct {
		M  string            `json:"m"`
		P  string            `json:"p"`
		H  map[string]string `json:"h"`
		B  any               `json:"b"`
		S  int               `json:"s"`
		J  map[string]any    `json:"j"`
		C  []string          `json:"c"`
		ML int64             `json:"ml"`
	}{normMethod(c.Method), c.Path, c.Headers, c.Body, c.ExpectStatus, c.ExpectJSON, c.ExpectBodyContains, c.MaxLatencyMs}
	b, _ := json.Marshal(key)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

func normMethod(m string) string {
	if m == "" {
		return "GET"
	}
	return string(bytesUpper(m))
}

func bytesUpper(s string) []byte {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return b
}

func (s *Store) path(name string) string { return filepath.Join(s.dir, name+".json") }

func (s *Store) load(name string) (*Suite, error) {
	if !ValidName(name) {
		return nil, errors.New("suite name must match [a-zA-Z0-9_-]{1,64}")
	}
	su := &Suite{Name: name}
	b, err := os.ReadFile(s.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return su, nil
	}
	if err != nil {
		return nil, err
	}
	return su, json.Unmarshal(b, su)
}

func (s *Store) save(su *Suite) error {
	b, _ := json.MarshalIndent(su, "", "  ")
	tmp := s.path(su.Name) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(su.Name))
}

func (s *Store) Get(name string) (*Suite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(name)
}

func (s *Store) Approve(name string, cases []apitest.TestCase, fromAI bool, rejected int) (*Suite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	su, err := s.load(name)
	if err != nil {
		return nil, err
	}
	have := map[string]bool{}
	for _, a := range su.Cases {
		have[a.Hash] = true
	}
	for _, c := range cases {
		h := Hash(c)
		if have[h] {
			continue
		}
		have[h] = true
		su.Cases = append(su.Cases, Approved{Hash: h, Case: c, ApprovedAt: time.Now().UTC(), FromAI: fromAI})
		if fromAI {
			su.AIAccept++
		}
	}
	su.AIReject += rejected
	sort.SliceStable(su.Cases, func(i, j int) bool { return su.Cases[i].ApprovedAt.Before(su.Cases[j].ApprovedAt) })
	return su, s.save(su)
}

func (s *Store) Remove(name, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	su, err := s.load(name)
	if err != nil {
		return err
	}
	kept := su.Cases[:0]
	found := false
	for _, a := range su.Cases {
		if a.Hash == hash {
			found = true
			continue
		}
		kept = append(kept, a)
	}
	if !found {
		return errors.New("case not found")
	}
	su.Cases = kept
	return s.save(su)
}

type Precision struct {
	Accepted   int     `json:"accepted"`
	Rejected   int     `json:"rejected"`
	Rate       float64 `json:"rate"`
	LowerBound float64 `json:"lower_bound_95"`
	Meets99    bool    `json:"meets_99_percent"`
}

func (su *Suite) Precision() Precision {
	n := su.AIAccept + su.AIReject
	p := Precision{Accepted: su.AIAccept, Rejected: su.AIReject}
	if n == 0 {
		return p
	}
	p.Rate = round3(float64(su.AIAccept) / float64(n))
	p.LowerBound = round3(wilsonLower(su.AIAccept, n))
	p.Meets99 = p.LowerBound >= 0.99
	return p
}

func wilsonLower(success, n int) float64 {
	const z = 1.96
	phat := float64(success) / float64(n)
	nn := float64(n)
	return (phat + z*z/(2*nn) - z*math.Sqrt((phat*(1-phat)+z*z/(4*nn))/nn)) / (1 + z*z/nn)
}

func round3(f float64) float64 { return math.Round(f*1000) / 1000 }
