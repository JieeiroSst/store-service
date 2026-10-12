package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/application"
)

type args map[string]any

func (a args) str(name string) (string, error) {
	v, ok := a[name]
	if !ok {
		return "", fmt.Errorf("Missing argument: %s", name)
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	case json.Number:
		return t.String(), nil
	}
	return "", fmt.Errorf("Wrong type of argument: %s", name)
}

func (a args) strList(names ...string) []string {
	for _, n := range names {
		switch t := a[n].(type) {
		case string:
			return []string{t}
		case []any:
			var out []string
			for _, v := range t {
				if s, ok := v.(string); ok {
					out = append(out, s)
				}
			}
			return out
		}
	}
	return nil
}

type rescode bool

type restuple string

type command func(ctx context.Context, a args) (any, error)

type commandSet struct {
	admin   *application.AdminService
	auth    *application.AuthService
	started time.Time
}

func (c *commandSet) table() map[string]command {
	return map[string]command{
		"status": func(context.Context, args) (any, error) {
			return restuple(c.admin.Status()), nil
		},
		"register": func(ctx context.Context, a args) (any, error) {
			user, host, password, err := three(a, "user", "host", "password")
			if err != nil {
				return nil, err
			}
			msg, err := c.admin.Register(ctx, user, host, password)
			if err != nil {
				return nil, err
			}
			return restuple(msg), nil
		},
		"unregister": func(ctx context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			if err := c.admin.Unregister(ctx, user, host); err != nil {
				return nil, err
			}
			return restuple(""), nil
		},
		"check_account": func(ctx context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.CheckAccount(ctx, user, host) == nil), nil
		},
		"check_password": func(ctx context.Context, a args) (any, error) {
			user, host, password, err := three(a, "user", "host", "password")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.CheckPassword(ctx, user, host, password) == nil), nil
		},
		"change_password": func(ctx context.Context, a args) (any, error) {
			user, host, newpass, err := three(a, "user", "host", "newpass")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.ChangePassword(ctx, user, host, newpass) == nil), nil
		},
		"registered_users": func(ctx context.Context, a args) (any, error) {
			host, err := a.str("host")
			if err != nil {
				return nil, err
			}
			users, err := c.admin.RegisteredUsers(ctx, host)
			if err != nil {
				return nil, err
			}
			return nonNil(users), nil
		},
		"registered_vhosts": func(context.Context, args) (any, error) {
			return c.admin.RegisteredVhosts(), nil
		},
		"connected_users": func(context.Context, args) (any, error) {
			return nonNil(c.admin.ConnectedUsers()), nil
		},
		"connected_users_number": func(context.Context, args) (any, error) {
			return c.admin.ConnectedUsersNumber(), nil
		},
		"connected_users_vhost": func(_ context.Context, a args) (any, error) {
			host, err := a.str("host")
			if err != nil {
				return nil, err
			}
			return nonNil(c.admin.ConnectedUsersVhost(host)), nil
		},
		"user_resources": func(_ context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			res, err := c.admin.UserResources(user, host)
			return nonNil(res), err
		},
		"num_resources": func(_ context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			res, err := c.admin.UserResources(user, host)
			return len(res), err
		},
		"kick_user": func(_ context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			return c.admin.KickUser(user, host)
		},
		"send_message": func(ctx context.Context, a args) (any, error) {
			typ, _ := a.str("type")
			from, to, err := two(a, "from", "to")
			if err != nil {
				return nil, err
			}
			subject, _ := a.str("subject")
			body, _ := a.str("body")
			return rescode(c.admin.SendMessage(ctx, typ, from, to, subject, body) == nil), nil
		},
		"send_stanza": func(ctx context.Context, a args) (any, error) {
			from, to, stanza, err := three(a, "from", "to", "stanza")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.SendStanza(ctx, from, to, stanza) == nil), nil
		},
		"get_roster": func(ctx context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			return c.admin.GetRoster(ctx, user, host)
		},
		"add_rosteritem": func(ctx context.Context, a args) (any, error) {
			lu, lh, err := two(a, "localuser", "localhost")
			if err != nil {
				return nil, err
			}
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			nick, _ := a.str("nick")
			subs, err := a.str("subs")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.AddRosterItem(ctx, lu, lh, user, host, nick, a.strList("groups", "group"), subs) == nil), nil
		},
		"delete_rosteritem": func(ctx context.Context, a args) (any, error) {
			lu, lh, err := two(a, "localuser", "localhost")
			if err != nil {
				return nil, err
			}
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			return rescode(c.admin.DeleteRosterItem(ctx, lu, lh, user, host) == nil), nil
		},
		"get_offline_count": func(ctx context.Context, a args) (any, error) {
			user, host, err := two(a, "user", "host")
			if err != nil {
				return nil, err
			}
			return c.admin.GetOfflineCount(ctx, user, host)
		},
		"delete_expired_messages": func(ctx context.Context, _ args) (any, error) {
			c.admin.DeleteExpiredMessages(ctx)
			return rescode(true), nil
		},
		"stats": func(ctx context.Context, a args) (any, error) {
			name, err := a.str("name")
			if err != nil {
				return nil, err
			}
			switch name {
			case "uptimeseconds":
				return int(time.Since(c.started).Seconds()), nil
			case "registeredusers":
				return c.admin.RegisteredUsersCount(ctx)
			case "onlineusers", "onlineusersnode":
				return c.admin.ConnectedUsersNumber(), nil
			}
			return nil, &application.CommandError{Status: 500, Code: 0, Message: "Unknown stat"}
		},
	}
}

func nonNil(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

func two(a args, n1, n2 string) (string, string, error) {
	v1, err := a.str(n1)
	if err != nil {
		return "", "", err
	}
	v2, err := a.str(n2)
	return v1, v2, err
}

func three(a args, n1, n2, n3 string) (string, string, string, error) {
	v1, v2, err := two(a, n1, n2)
	if err != nil {
		return "", "", "", err
	}
	v3, err := a.str(n3)
	return v1, v2, v3, err
}

func asCommandError(err error) (*application.CommandError, bool) {
	var ce *application.CommandError
	return ce, errors.As(err, &ce)
}
