package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

var snapshotID = regexp.MustCompile(`^[0-9a-f]{32}$`)

func newSnapshotID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) rank(
	ctx context.Context, scope string, q port.PageQuery,
	compute func(m *engine.Model) ([]engine.Scored, error),
) (*port.Page, error) {
	if q.Snapshot != "" && !snapshotID.MatchString(q.Snapshot) {
		return nil, fmt.Errorf("%w: bad snapshot", port.ErrInvalid)
	}
	m := s.model.Load()

	if q.Snapshot != "" {
		ranked, ok, err := s.snapshots.Load(ctx, q.Snapshot, scope, s.now())
		if err != nil {
			log.Printf("load snapshot: %v", err)
		} else if ok {
			return s.paginate(m, ranked, q, q.Snapshot), nil
		}
	}

	ranked, err := compute(m)
	if err != nil {
		return nil, err
	}
	token := ""
	if len(ranked) > pageSizeOf(q) {
		if id, err := newSnapshotID(); err != nil {
			log.Printf("snapshot id: %v", err)
		} else if err := s.snapshots.Save(ctx, id, scope, ranked, s.now().Add(s.cfg.Snapshot.TTL)); err != nil {
			log.Printf("save snapshot: %v", err)
		} else {
			token = id
		}
	}
	return s.paginate(m, ranked, q, token), nil
}
