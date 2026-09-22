package infrastructure

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/ekyc-service/config"
	httpadapter "github.com/JIeeiroSst/ekyc-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/cardreader"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/facebio"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/nfcreader"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/pyextract"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/storage"
	"github.com/JIeeiroSst/ekyc-service/internal/adapter/secondary/userclient"
	"github.com/JIeeiroSst/ekyc-service/internal/application"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
	"github.com/JIeeiroSst/ekyc-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/ekyc-service/internal/infrastructure/server"
)

func newCardReader(cfg *config.Config) port.CardReader {
	if cfg.Providers.CardReader == "python" {
		return pyextract.NewCardReader(pyextract.NewClient(cfg.Python.Bin, cfg.Python.ScriptPath, cfg.Python.Timeout()))
	}
	return cardreader.NewCardReader()
}

func newFaceAnalyzer(cfg *config.Config) port.FaceAnalyzer {
	if cfg.Providers.FaceAnalyzer == "python" {
		return pyextract.NewFaceAnalyzer(pyextract.NewClient(cfg.Python.Bin, cfg.Python.ScriptPath, cfg.Python.Timeout()))
	}
	return facebio.NewFaceAnalyzer()
}

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module,
	storage.Module,

	repository.Module,
	fx.Provide(newCardReader),
	fx.Provide(newFaceAnalyzer),
	nfcreader.Module,
	userclient.Module,

	application.Module,

	httpadapter.Module,

	fx.Invoke(server.New),
)
