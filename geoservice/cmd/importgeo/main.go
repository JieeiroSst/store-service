package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/geoservice/config"
	"github.com/geoservice/internal/adapter/importer"
	"github.com/geoservice/internal/adapter/migration"
	"github.com/geoservice/internal/adapter/repository"
	"github.com/geoservice/internal/core/domain"
)

func main() {
	var (
		l1 = flag.String("l1", "", "GADM level-1 GeoJSON (provinces)")
		l2 = flag.String("l2", "", "GADM level-2 GeoJSON (districts)")
		l3 = flag.String("l3", "", "GADM level-3 GeoJSON (wards)")
	)
	flag.Parse()

	if *l1 == "" && *l2 == "" && *l3 == "" {
		fmt.Fprintln(os.Stderr, "error: provide at least one of -l1 / -l2 / -l3")
		flag.Usage()
		os.Exit(2)
	}

	cfg, err := config.Load()
	must(err)
	log, _ := zap.NewDevelopment()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	must(err)
	defer pool.Close()
	must(pool.Ping(ctx))
	must(migration.Run(ctx, pool, log))

	repo := repository.NewGeoRepository(pool)
	im := importer.New(repo, log)

	for _, step := range []struct {
		file  string
		level domain.AdminLevel
	}{
		{*l1, domain.LevelProvince},
		{*l2, domain.LevelDistrict},
		{*l3, domain.LevelWard},
	} {
		if step.file == "" {
			continue
		}
		m, t, err := im.ImportFile(ctx, step.file, step.level)
		must(err)
		log.Info("level done", zap.Int("level", int(step.level)),
			zap.Int("matched", m), zap.Int("total", t))
	}

	p, d, w, err := repo.CountSeeded(ctx)
	must(err)
	log.Info("seeded counts", zap.Int64("provinces", p), zap.Int64("districts", d), zap.Int64("wards", w))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
