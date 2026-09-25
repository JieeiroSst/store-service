package engine

import (
	"math"
	"sort"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
)

type Config struct {
	HalfLife                                                   time.Duration
	MaxItemsPerUser                                            int
	Neighbors                                                  int
	NewHalfLife                                                time.Duration
	Diversity                                                  float64
	DiversifyTop                                               int
	CFWeight, ContentWeight, PopularityWeight, FreshnessWeight float64
}

func DefaultConfig() Config {
	return Config{
		HalfLife:         7 * 24 * time.Hour,
		MaxItemsPerUser:  100,
		Neighbors:        50,
		NewHalfLife:      14 * 24 * time.Hour,
		Diversity:        0.25,
		DiversifyTop:     100,
		CFWeight:         0.50,
		ContentWeight:    0.28,
		PopularityWeight: 0.12,
		FreshnessWeight:  0.10,
	}
}

type Scored struct {
	VideoID   string
	Score     float64
	Reason    string
	BecauseID string
}

type neighbor struct {
	idx int
	sim float64
}

type Model struct {
	cfg      Config
	videos   []model.Video
	index    map[string]int
	vecs     []vector
	pop      []float64
	fresh    []float64
	trending []int
	nbrs     [][]neighbor
}

func (m *Model) Len() int { return len(m.videos) }

func (m *Model) Video(id string) (model.Video, bool) {
	i, ok := m.index[id]
	if !ok {
		return model.Video{}, false
	}
	return m.videos[i], true
}

func Build(videos []model.Video, interactions []model.Interaction, now time.Time, cfg Config) *Model {
	m := &Model{cfg: cfg, index: map[string]int{}}
	for _, v := range videos {
		if !v.Ready() {
			continue
		}
		if _, dup := m.index[v.ID]; dup {
			continue
		}
		m.index[v.ID] = len(m.videos)
		m.videos = append(m.videos, v)
	}
	m.buildVectors()
	m.buildPopularity(interactions, now)
	m.buildFreshness(now)
	m.buildNeighbors(interactions)
	return m
}

func (m *Model) buildVectors() {
	n := len(m.videos)
	tfs := make([]vector, n)
	df := map[string]int{}
	for i, v := range m.videos {
		tf := vector{}
		for _, t := range tokenize(v.Title) {
			tf[t] += 2
		}
		for _, tag := range v.Tags {
			for _, t := range tokenize(tag) {
				tf[t] += 3
			}
		}
		for _, t := range tokenize(v.Description) {
			tf[t]++
		}
		for t := range tf {
			df[t]++
		}
		tfs[i] = tf
	}
	m.vecs = tfs
	for _, tf := range tfs {
		for t, c := range tf {
			idf := math.Log(1 + float64(n)/float64(1+df[t]))
			tf[t] = (1 + math.Log(c)) * idf
		}
		tf.normalize()
	}
}

func (m *Model) buildPopularity(interactions []model.Interaction, now time.Time) {
	n := len(m.videos)
	raw := make([]float64, n)
	for i, v := range m.videos {
		raw[i] = 0.25 * math.Log1p(float64(v.Views))
	}
	for _, in := range interactions {
		i, ok := m.index[in.VideoID]
		if !ok {
			continue
		}
		w := in.Weight()
		if w <= 0 {
			continue
		}
		age := now.Sub(in.At)
		if age < 0 {
			age = 0
		}
		raw[i] += w * math.Exp2(-float64(age)/float64(m.cfg.HalfLife))
	}
	m.pop = normalizeMax(raw)
	m.trending = make([]int, n)
	for i := range m.trending {
		m.trending[i] = i
	}
	sort.SliceStable(m.trending, func(a, b int) bool { return m.pop[m.trending[a]] > m.pop[m.trending[b]] })
}

func (m *Model) buildNeighbors(interactions []model.Interaction) {
	type entry struct {
		idx int
		w   float64
	}
	perUser := map[string]map[int]float64{}
	for _, in := range interactions {
		i, ok := m.index[in.VideoID]
		if !ok {
			continue
		}
		u := perUser[in.UserID]
		if u == nil {
			u = map[int]float64{}
			perUser[in.UserID] = u
		}
		u[i] += in.Weight()
	}

	norm := make([]float64, len(m.videos))
	co := make([]map[int]float64, len(m.videos))
	for _, items := range perUser {
		var es []entry
		for i, w := range items {
			if w > 0 {
				es = append(es, entry{i, w})
			}
		}
		if len(es) < 2 {
			for _, e := range es {
				norm[e.idx] += e.w * e.w
			}
			continue
		}
		sort.Slice(es, func(a, b int) bool {
			if es[a].w != es[b].w {
				return es[a].w > es[b].w
			}
			return es[a].idx < es[b].idx
		})
		if len(es) > m.cfg.MaxItemsPerUser {
			es = es[:m.cfg.MaxItemsPerUser]
		}
		damp := 1 / math.Log(2+float64(len(es)))
		for _, a := range es {
			norm[a.idx] += a.w * a.w
			for _, b := range es {
				if a.idx == b.idx {
					continue
				}
				if co[a.idx] == nil {
					co[a.idx] = map[int]float64{}
				}
				co[a.idx][b.idx] += damp * a.w * b.w
			}
		}
	}

	m.nbrs = make([][]neighbor, len(m.videos))
	for i, row := range co {
		ns := make([]neighbor, 0, len(row))
		for j, c := range row {
			if s := c / math.Sqrt(norm[i]*norm[j]); s > 0 {
				ns = append(ns, neighbor{j, math.Min(s, 1)})
			}
		}
		sort.Slice(ns, func(a, b int) bool {
			if ns[a].sim != ns[b].sim {
				return ns[a].sim > ns[b].sim
			}
			return ns[a].idx < ns[b].idx
		})
		if len(ns) > m.cfg.Neighbors {
			ns = ns[:m.cfg.Neighbors]
		}
		m.nbrs[i] = ns
	}
}

func (m *Model) buildFreshness(now time.Time) {
	m.fresh = make([]float64, len(m.videos))
	for i, v := range m.videos {
		if v.CreatedAt.IsZero() {
			continue
		}
		age := max(now.Sub(v.CreatedAt), 0)
		m.fresh[i] = math.Exp2(-float64(age) / float64(m.cfg.NewHalfLife))
	}
}

func (m *Model) NewReleases(limit int, exclude map[string]bool) []Scored {
	var out []Scored
	for i, v := range m.videos {
		if exclude[v.ID] || v.CreatedAt.IsZero() {
			continue
		}
		out = append(out, Scored{VideoID: v.ID, Score: 0.8*m.fresh[i] + 0.2*m.pop[i], Reason: model.ReasonNewRelease})
	}
	return top(out, limit)
}

func (m *Model) diversify(s []Scored) []Scored {
	n := min(len(s), m.cfg.DiversifyTop)
	if m.cfg.Diversity <= 0 || n < 3 {
		return s
	}
	head := append([]Scored(nil), s[:n]...)
	maxSim := make([]float64, n)
	picked := make([]Scored, 0, len(s))
	last := -1
	for len(head) > 0 {
		best, bestVal := 0, math.Inf(-1)
		for i, c := range head {
			if last >= 0 {
				maxSim[i] = math.Max(maxSim[i], dot(m.vecs[m.index[c.VideoID]], m.vecs[last]))
			}
			if v := c.Score - m.cfg.Diversity*maxSim[i]; v > bestVal {
				best, bestVal = i, v
			}
		}
		last = m.index[head[best].VideoID]
		picked = append(picked, head[best])
		head = append(head[:best], head[best+1:]...)
		maxSim = append(maxSim[:best], maxSim[best+1:]...)
	}
	return append(picked, s[n:]...)
}

func (m *Model) Trending(limit int, exclude map[string]bool) []Scored {
	var out []Scored
	for _, i := range m.trending {
		v := m.videos[i]
		if exclude[v.ID] {
			continue
		}
		out = append(out, Scored{VideoID: v.ID, Score: m.pop[i], Reason: model.ReasonTrending})
		if len(out) == limit {
			break
		}
	}
	return out
}

func (m *Model) Similar(id string, limit int) ([]Scored, bool) {
	src, ok := m.index[id]
	if !ok {
		return nil, false
	}
	cf := make([]float64, len(m.videos))
	for _, n := range m.nbrs[src] {
		cf[n.idx] = n.sim
	}
	var out []Scored
	for j := range m.videos {
		if j == src {
			continue
		}
		content := dot(m.vecs[src], m.vecs[j])
		score := 0.6*cf[j] + 0.4*content
		if score <= 0 {
			continue
		}
		reason := model.ReasonSimilarContent
		if 0.6*cf[j] >= 0.4*content {
			reason = model.ReasonWatchedTogether
		}
		out = append(out, Scored{VideoID: m.videos[j].ID, Score: score, Reason: reason, BecauseID: id})
	}
	return trim(m.diversify(top(out, -1)), limit), true
}

func (m *Model) ForUser(history []model.Interaction, limit int) []Scored {
	weights := map[int]float64{}
	seen := map[string]bool{}
	for _, in := range history {
		seen[in.VideoID] = true
		if i, ok := m.index[in.VideoID]; ok {
			weights[i] += in.Weight()
		}
	}
	positive := false
	for _, w := range weights {
		if w > 0 {
			positive = true
			break
		}
	}
	if !positive {
		return m.Trending(limit, seen)
	}

	cf := make([]float64, len(m.videos))
	bestSrc := make([]int, len(m.videos))
	bestContribution := make([]float64, len(m.videos))
	for i := range bestSrc {
		bestSrc[i] = -1
	}
	profile := vector{}
	for i, w := range weights {
		for _, n := range m.nbrs[i] {
			c := w * n.sim
			cf[n.idx] += c
			if c > bestContribution[n.idx] {
				bestContribution[n.idx], bestSrc[n.idx] = c, i
			}
		}
		f := w
		if f < 0 {
			f *= 0.5
		}
		for t, x := range m.vecs[i] {
			profile[t] += f * x
		}
	}
	profile.normalize()
	cfNorm := normalizeMax(cf)

	var out []Scored
	for j, v := range m.videos {
		if seen[v.ID] {
			continue
		}
		content := math.Max(0, dot(profile, m.vecs[j]))
		cfScore := math.Max(0, cfNorm[j])
		score := m.cfg.CFWeight*cfScore + m.cfg.ContentWeight*content +
			m.cfg.PopularityWeight*m.pop[j] + m.cfg.FreshnessWeight*m.fresh[j]
		if score <= 0 {
			continue
		}
		s := Scored{VideoID: v.ID, Score: score, Reason: model.ReasonForYou}
		if bestSrc[j] >= 0 && m.cfg.CFWeight*cfScore >= m.cfg.ContentWeight*content {
			s.Reason, s.BecauseID = model.ReasonBecauseWatched, m.videos[bestSrc[j]].ID
		}
		out = append(out, s)
	}
	out = trim(m.diversify(top(out, -1)), limit)
	if len(out) < limit {
		have := map[string]bool{}
		for _, s := range out {
			have[s.VideoID] = true
		}
		for id := range seen {
			have[id] = true
		}
		out = append(out, m.Trending(limit-len(out), have)...)
	}
	return out
}

func top(s []Scored, limit int) []Scored {
	sort.Slice(s, func(a, b int) bool {
		if s[a].Score != s[b].Score {
			return s[a].Score > s[b].Score
		}
		return s[a].VideoID < s[b].VideoID
	})
	return trim(s, limit)
}

func trim(s []Scored, limit int) []Scored {
	if limit >= 0 && len(s) > limit {
		s = s[:limit]
	}
	return s
}

func normalizeMax(x []float64) []float64 {
	var max float64
	for _, v := range x {
		if v > max {
			max = v
		}
	}
	out := make([]float64, len(x))
	if max == 0 {
		return out
	}
	for i, v := range x {
		out[i] = v / max
	}
	return out
}
