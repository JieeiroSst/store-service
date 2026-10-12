package application

import (
	"context"

	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewIQRegistry,
		NewSessionManager,
		NewRouter,
		NewAuthService,
		NewRosterService,
		NewOfflineService,
		NewC2SService,
		NewModules,
		NewAdminService,
		NewAdminBootstrap,
		NewHTTPUploadService,
	),
	fx.Invoke(wireHooks),
	fx.Invoke(startAdminBootstrap),
)

func wireHooks(r *Router, auth *AuthService, roster *RosterService, offline *OfflineService, sm *SessionManager, mods *Modules, upload *HTTPUploadService) {
	upload.Register()
	r.Use(roster, offline, auth)
	auth.OnRemoveUser(roster.RemoveUser)
	auth.OnRemoveUser(offline.RemoveUser)
	auth.OnRemoveUser(func(_ context.Context, luser, lserver string) {
		sm.DeliverToUser(luser, lserver, port.SessionEvent{Kind: port.EventExit, Reason: "User removed"})
	})
	mods.Register()
}
