package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) XMPPRegister(ctx context.Context, user string, host string, password string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/register", Body: map[string]any{"user": user, "host": host, "password": password}}, "")
}

func (r *mutationResolver) XMPPUnregister(ctx context.Context, user string, host string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/unregister", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *mutationResolver) XMPPChangePassword(ctx context.Context, user string, host string, newpass string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/change_password", Body: map[string]any{"user": user, "host": host, "newpass": newpass}}, "")
}

func (r *mutationResolver) XMPPKickUser(ctx context.Context, user string, host string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/kick_user", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *mutationResolver) XMPPSendMessage(ctx context.Context, typeArg string, from string, to string, subject *string, body string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/send_message", Body: map[string]any{"type": typeArg, "from": from, "to": to, "subject": subject, "body": body}}, "")
}

func (r *mutationResolver) XMPPSendStanza(ctx context.Context, from string, to string, stanza string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/send_stanza", Body: map[string]any{"from": from, "to": to, "stanza": stanza}}, "")
}

func (r *mutationResolver) XMPPAddRosterItem(ctx context.Context, localuser string, localhost string, user string, host string, nick *string, group *string, subs *string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/add_rosteritem", Body: map[string]any{"localuser": localuser, "localhost": localhost, "user": user, "host": host, "nick": nick, "group": group, "subs": subs}}, "")
}

func (r *mutationResolver) XMPPDeleteRosterItem(ctx context.Context, localuser string, localhost string, user string, host string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/delete_rosteritem", Body: map[string]any{"localuser": localuser, "localhost": localhost, "user": user, "host": host}}, "")
}

func (r *mutationResolver) XMPPDeleteExpiredMessages(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/delete_expired_messages", Body: map[string]any{}}, "")
}

func (r *queryResolver) XMPPStatus(ctx context.Context) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/status", Body: map[string]any{}}, "")
}

func (r *queryResolver) XMPPCheckAccount(ctx context.Context, user string, host string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/check_account", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *queryResolver) XMPPRegisteredUsers(ctx context.Context, host string) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/registered_users", Body: map[string]any{"host": host}}, "")
}

func (r *queryResolver) XMPPRegisteredVhosts(ctx context.Context) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/registered_vhosts", Body: map[string]any{}}, "")
}

func (r *queryResolver) XMPPConnectedUsers(ctx context.Context) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/connected_users", Body: map[string]any{}}, "")
}

func (r *queryResolver) XMPPConnectedUsersNumber(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/connected_users_number", Body: map[string]any{}}, "")
}

func (r *queryResolver) XMPPConnectedUsersVhost(ctx context.Context, host string) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/connected_users_vhost", Body: map[string]any{"host": host}}, "")
}

func (r *queryResolver) XMPPUserResources(ctx context.Context, user string, host string) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/user_resources", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *queryResolver) XMPPNumResources(ctx context.Context, user string, host string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/num_resources", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *queryResolver) XMPPRoster(ctx context.Context, user string, host string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/get_roster", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *queryResolver) XMPPOfflineCount(ctx context.Context, user string, host string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/get_offline_count", Body: map[string]any{"user": user, "host": host}}, "")
}

func (r *queryResolver) XMPPStats(ctx context.Context, name string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: "/api/stats", Body: map[string]any{"name": name}}, "")
}

func (r *queryResolver) XMPPCommand(ctx context.Context, command string, args map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ejabberd-service", rest.Request{Method: "POST", Path: rest.P("/api/%v", command), Body: args}, "")
}
