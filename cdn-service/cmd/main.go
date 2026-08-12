package main

import (
	httpadapter "github.com/JIeeiroSst/cdn-service/internal/adapters/http"
	"github.com/JIeeiroSst/cdn-service/internal/adapters/repository"
	"github.com/JIeeiroSst/cdn-service/internal/adapters/storage"
	"github.com/JIeeiroSst/cdn-service/internal/config"
	"github.com/JIeeiroSst/cdn-service/internal/domain"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		repository.Module,
		storage.Module,
		domain.Module,
		httpadapter.Module,
	).Run()
}
