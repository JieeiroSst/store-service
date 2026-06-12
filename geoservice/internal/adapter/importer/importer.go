package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/geoservice/internal/core/domain"
	"github.com/geoservice/internal/core/port"
)

type featureCollection struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	Properties map[string]any  `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type Importer struct {
	repo port.GeoRepository
	log  *zap.Logger
}

func New(repo port.GeoRepository, log *zap.Logger) *Importer {
	return &Importer{repo: repo, log: log}
}

func (im *Importer) ImportFile(ctx context.Context, path string, level domain.AdminLevel) (matched, total int, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("read %s: %w", path, err)
	}
	var fc featureCollection
	if err := json.Unmarshal(raw, &fc); err != nil {
		return 0, 0, fmt.Errorf("parse geojson: %w", err)
	}
	if !strings.EqualFold(fc.Type, "FeatureCollection") {
		return 0, 0, fmt.Errorf("expected FeatureCollection, got %q", fc.Type)
	}

	for i, f := range fc.Features {
		feat, ok := im.toFeature(level, f, i)
		if !ok {
			continue
		}
		total++
		ok, err := im.repo.AttachGeometryByName(ctx, feat)
		if err != nil {
			return matched, total, err
		}
		if ok {
			matched++
		} else {
			im.log.Warn("no seeded row matched",
				zap.Int("level", int(level)), zap.String("name", feat.Name), zap.String("gid", feat.Gid))
		}
	}
	im.log.Info("attach complete",
		zap.Int("level", int(level)), zap.Int("matched", matched), zap.Int("total", total), zap.String("file", path))
	return matched, total, nil
}

func (im *Importer) toFeature(level domain.AdminLevel, f feature, idx int) (domain.GeoFeature, bool) {
	p := f.Properties
	gf := domain.GeoFeature{Level: level, GeoJSON: string(f.Geometry)}

	switch level {
	case domain.LevelProvince:
		gf.Gid = str(p, "GID_1")
		gf.Name = first(str(p, "NAME_1"), str(p, "VARNAME_1"))
	case domain.LevelDistrict:
		gf.Gid = str(p, "GID_2")
		gf.ProvinceGid = str(p, "GID_1")
		gf.Name = str(p, "NAME_2")
	case domain.LevelWard:
		gf.Gid = str(p, "GID_3")
		gf.DistrictGid = str(p, "GID_2")
		gf.ProvinceGid = str(p, "GID_1")
		gf.Name = str(p, "NAME_3")
	}

	if gf.Name == "" {
		im.log.Warn("skip feature: empty name", zap.Int("index", idx))
		return gf, false
	}
	if len(gf.GeoJSON) == 0 || gf.GeoJSON == "null" {
		im.log.Warn("skip feature: no geometry", zap.String("name", gf.Name))
		return gf, false
	}
	return gf, true
}

func str(p map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := p[k]; ok {
			if s := asStr(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func asStr(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSuffix(fmt.Sprintf("%v", t), ".0")
	default:
		return ""
	}
}
