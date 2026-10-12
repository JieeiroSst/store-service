package postgres

import (
	"cmp"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"github.com/lib/pq"
)

var (
	//go:embed schema.sql
	schema string
	//go:embed users_schema.sql
	usersSchema string
)

type DB struct {
	*sql.DB
}

func Open(cfg *config.Config) (*DB, error) {
	db, err := sql.Open("postgres", cfg.SQL.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.SQL.PoolSize)
	db.SetMaxIdleConns(cfg.SQL.PoolSize)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return &DB{DB: db}, nil
}

func (d *DB) Migrate(ctx context.Context, withUsers bool) error {
	var lastErr error
	for attempt := 1; attempt <= 30; attempt++ {
		if lastErr = d.PingContext(ctx); lastErr == nil {
			break
		}
		slog.Warn("postgres not ready", "attempt", attempt, "err", lastErr)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	if lastErr != nil {
		return fmt.Errorf("connect postgres: %w", lastErr)
	}
	if _, err := d.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply sql schema: %w", err)
	}
	if withUsers {
		if _, err := d.ExecContext(ctx, usersSchema); err != nil {
			return fmt.Errorf("apply users schema: %w", err)
		}
	}
	return nil
}

func (d *DB) Ping(ctx context.Context) error {
	return d.PingContext(ctx)
}

type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type UserStore struct {
	db *DB
}

func NewUserStore(db *DB) *UserStore { return &UserStore{db: db} }

var _ port.UserRepository = (*UserStore)(nil)

func encodePassword(p domain.Password) (pass, serverKey, salt string, iter int) {
	if !p.IsScram() {
		return p.Plain, "", "", 0
	}
	s := p.Scram
	stored := s.StoredKey
	switch s.Hash {
	case domain.ScramSHA256:
		stored = "sha256:" + stored
	case domain.ScramSHA512:
		stored = "sha512:" + stored
	}
	return stored, s.ServerKey, s.Salt, s.IterationCount
}

func decodePassword(pass, serverKey, salt string, iter int) domain.Password {
	if serverKey == "" {
		return domain.Password{Plain: pass}
	}
	h, stored := domain.ScramSHA1, pass
	switch {
	case strings.HasPrefix(pass, "sha256:"):
		h, stored = domain.ScramSHA256, strings.TrimPrefix(pass, "sha256:")
	case strings.HasPrefix(pass, "sha512:"):
		h, stored = domain.ScramSHA512, strings.TrimPrefix(pass, "sha512:")
	}
	return domain.Password{Scram: &domain.Scram{StoredKey: stored, ServerKey: serverKey, Salt: salt, Hash: h, IterationCount: iter}}
}

func (s *UserStore) GetPassword(ctx context.Context, luser, lserver string) (domain.Password, error) {
	var pass, serverKey, salt string
	var iter int
	err := s.db.QueryRowContext(ctx,
		`SELECT password, serverkey, salt, iterationcount FROM users WHERE username=$1 AND server_host=$2`,
		luser, lserver).Scan(&pass, &serverKey, &salt, &iter)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Password{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.Password{}, err
	}
	return decodePassword(pass, serverKey, salt, iter), nil
}

func (s *UserStore) SetPassword(ctx context.Context, luser, lserver string, p domain.Password) error {
	pass, serverKey, salt, iter := encodePassword(p)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, server_host, password, serverkey, salt, iterationcount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (server_host, username) DO UPDATE
		SET password=EXCLUDED.password, serverkey=EXCLUDED.serverkey, salt=EXCLUDED.salt, iterationcount=EXCLUDED.iterationcount`,
		luser, lserver, pass, serverKey, salt, iter)
	return err
}

func (s *UserStore) TryRegister(ctx context.Context, luser, lserver string, p domain.Password) error {
	pass, serverKey, salt, iter := encodePassword(p)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, server_host, password, serverkey, salt, iterationcount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (server_host, username) DO NOTHING`,
		luser, lserver, pass, serverKey, salt, iter)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrUserExists
	}
	return nil
}

func (s *UserStore) UserExists(ctx context.Context, luser, lserver string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE username=$1 AND server_host=$2`, luser, lserver).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *UserStore) RemoveUser(ctx context.Context, luser, lserver string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE username=$1 AND server_host=$2`, luser, lserver)
	return err
}

func (s *UserStore) ListUsers(ctx context.Context, lserver string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT username FROM users WHERE server_host=$1 ORDER BY username`, lserver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *UserStore) CountUsers(ctx context.Context, lserver string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE server_host=$1`, lserver).Scan(&n)
	return n, err
}

type RosterStore struct {
	db *DB
	q  querier
}

func NewRosterStore(db *DB) *RosterStore { return &RosterStore{db: db, q: db} }

var _ port.RosterRepository = (*RosterStore)(nil)

func encodeSubscription(s domain.Subscription) string {
	switch s {
	case domain.SubBoth:
		return "B"
	case domain.SubTo:
		return "T"
	case domain.SubFrom:
		return "F"
	}
	return "N"
}

func decodeSubscription(s string) domain.Subscription {
	switch strings.TrimSpace(s) {
	case "B":
		return domain.SubBoth
	case "T":
		return domain.SubTo
	case "F":
		return domain.SubFrom
	}
	return domain.SubNone
}

func encodeAsk(a domain.Ask) string {
	switch a {
	case domain.AskSubscribe:
		return "S"
	case domain.AskUnsubscribe:
		return "U"
	case domain.AskBoth:
		return "B"
	case domain.AskOut:
		return "O"
	case domain.AskIn:
		return "I"
	}
	return "N"
}

func decodeAsk(a string) domain.Ask {
	switch strings.TrimSpace(a) {
	case "S":
		return domain.AskSubscribe
	case "U":
		return domain.AskUnsubscribe
	case "B":
		return domain.AskBoth
	case "O":
		return domain.AskOut
	case "I":
		return domain.AskIn
	}
	return domain.AskNone
}

func (s *RosterStore) groups(ctx context.Context, luser, lserver, jid string) ([]string, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT grp FROM rostergroups WHERE username=$1 AND server_host=$2 AND jid=$3`, luser, lserver, jid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *RosterStore) scanItems(ctx context.Context, luser, lserver string, rows *sql.Rows) ([]domain.RosterItem, error) {
	defer rows.Close()
	var out []domain.RosterItem
	for rows.Next() {
		var jidS, nick, sub, ask, askMsg string
		if err := rows.Scan(&jidS, &nick, &sub, &ask, &askMsg); err != nil {
			return nil, err
		}
		j, err := domain.ParseJID(jidS)
		if err != nil {
			slog.Error("invalid jid in rosterusers", "user", luser, "server", lserver, "jid", jidS)
			continue
		}
		out = append(out, domain.RosterItem{
			LUser: luser, LServer: lserver, JID: j.LJID(), Name: nick,
			Subscription: decodeSubscription(sub), Ask: decodeAsk(ask), AskMessage: askMsg,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		g, err := s.groups(ctx, luser, lserver, out[i].JID.String())
		if err != nil {
			return nil, err
		}
		out[i].Groups = g
	}
	return out, nil
}

func (s *RosterStore) GetRoster(ctx context.Context, luser, lserver string) ([]domain.RosterItem, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT jid, nick, subscription, ask, askmessage FROM rosterusers WHERE username=$1 AND server_host=$2 ORDER BY jid`,
		luser, lserver)
	if err != nil {
		return nil, err
	}
	return s.scanItems(ctx, luser, lserver, rows)
}

func (s *RosterStore) GetRosterItem(ctx context.Context, luser, lserver string, jid domain.LJID) (domain.RosterItem, bool, error) {
	rows, err := s.q.QueryContext(ctx,
		`SELECT jid, nick, subscription, ask, askmessage FROM rosterusers WHERE username=$1 AND server_host=$2 AND jid=$3`,
		luser, lserver, jid.String())
	if err != nil {
		return domain.RosterItem{}, false, err
	}
	items, err := s.scanItems(ctx, luser, lserver, rows)
	if err != nil || len(items) == 0 {
		return domain.RosterItem{}, false, err
	}
	return items[0], true, nil
}

func (s *RosterStore) UpdateRoster(ctx context.Context, item domain.RosterItem) error {
	jid := item.JID.String()
	if _, err := s.q.ExecContext(ctx, `
		INSERT INTO rosterusers (username, server_host, jid, nick, subscription, ask, askmessage, server, subscribe, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'N', '', 'item')
		ON CONFLICT (server_host, username, jid) DO UPDATE
		SET nick=EXCLUDED.nick, subscription=EXCLUDED.subscription, ask=EXCLUDED.ask, askmessage=EXCLUDED.askmessage`,
		item.LUser, item.LServer, jid, item.Name, encodeSubscription(item.Subscription), encodeAsk(item.Ask), item.AskMessage); err != nil {
		return err
	}
	if _, err := s.q.ExecContext(ctx,
		`DELETE FROM rostergroups WHERE username=$1 AND server_host=$2 AND jid=$3`, item.LUser, item.LServer, jid); err != nil {
		return err
	}
	for _, g := range item.Groups {
		if _, err := s.q.ExecContext(ctx,
			`INSERT INTO rostergroups (username, server_host, jid, grp) VALUES ($1, $2, $3, $4)`,
			item.LUser, item.LServer, jid, g); err != nil {
			return err
		}
	}
	return nil
}

func (s *RosterStore) DeleteRoster(ctx context.Context, luser, lserver string, jid domain.LJID) error {
	j := jid.String()
	if _, err := s.q.ExecContext(ctx,
		`DELETE FROM rosterusers WHERE username=$1 AND server_host=$2 AND jid=$3`, luser, lserver, j); err != nil {
		return err
	}
	_, err := s.q.ExecContext(ctx,
		`DELETE FROM rostergroups WHERE username=$1 AND server_host=$2 AND jid=$3`, luser, lserver, j)
	return err
}

func (s *RosterStore) RemoveUser(ctx context.Context, luser, lserver string) error {
	for _, q := range []string{
		`DELETE FROM rosterusers WHERE username=$1 AND server_host=$2`,
		`DELETE FROM rostergroups WHERE username=$1 AND server_host=$2`,
		`DELETE FROM roster_version WHERE username=$1 AND server_host=$2`,
	} {
		if _, err := s.q.ExecContext(ctx, q, luser, lserver); err != nil {
			return err
		}
	}
	return nil
}

func (s *RosterStore) Transaction(ctx context.Context, luser, lserver string, fn func(tx port.RosterRepository) error) error {
	if _, ok := s.q.(*sql.Tx); ok {
		return fn(s)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, luser+"@"+lserver); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := fn(&RosterStore{db: s.db, q: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type OfflineStore struct {
	db *DB
}

func NewOfflineStore(db *DB) *OfflineStore { return &OfflineStore{db: db} }

var _ port.OfflineRepository = (*OfflineStore)(nil)

func (s *OfflineStore) Store(ctx context.Context, msg domain.OfflineMessage) error {
	pkt := msg.Packet.Copy()
	domain.SetFromTo(pkt, msg.From, msg.To)
	host, _ := domain.NewJID("", msg.LServer, "")
	domain.AddDelay(pkt, host, msg.Timestamp, "Offline Storage")
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO spool (username, server_host, xml) VALUES ($1, $2, $3)`,
		msg.LUser, msg.LServer, pkt.StringNS(domain.NSClient))
	return err
}

func decodeOffline(luser, lserver, xml string) (domain.OfflineMessage, bool) {
	el, err := domain.ParseElement(xml)
	if err != nil {
		slog.Error("failed to decode offline packet", "user", luser, "server", lserver, "err", err)
		return domain.OfflineMessage{}, false
	}
	if el.Namespace() == "" {
		el.SetAttr("xmlns", domain.NSClient)
	}
	from, _ := domain.StanzaFrom(el)
	to, _ := domain.StanzaTo(el)
	ts := time.Now()
	out := el.Children[:0]
	for _, c := range el.Children {
		if d, ok := c.(*domain.Element); ok && d.Name == "delay" && d.Namespace() == domain.NSDelay && d.Attr("from") == lserver {
			if t, err := time.Parse(time.RFC3339Nano, d.Attr("stamp")); err == nil {
				ts = t
			}
			continue
		}
		out = append(out, c)
	}
	el.Children = out
	return domain.OfflineMessage{
		LUser: luser, LServer: lserver, Timestamp: ts, Expire: domain.FindExpire(ts, el),
		From: from, To: to, Packet: el,
	}, true
}

func (s *OfflineStore) Pop(ctx context.Context, luser, lserver string) ([]domain.OfflineMessage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx,
		`DELETE FROM spool WHERE username=$1 AND server_host=$2 RETURNING xml, seq`, luser, lserver)
	if err != nil {
		return nil, err
	}
	type row struct {
		xml string
		seq int64
	}
	var raw []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.xml, &r.seq); err != nil {
			rows.Close()
			return nil, err
		}
		raw = append(raw, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slices.SortFunc(raw, func(a, b row) int { return cmp.Compare(a.seq, b.seq) })
	var out []domain.OfflineMessage
	for _, r := range raw {
		if m, ok := decodeOffline(luser, lserver, r.xml); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *OfflineStore) Count(ctx context.Context, luser, lserver string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM spool WHERE username=$1 AND server_host=$2`, luser, lserver).Scan(&n)
	return n, err
}

func (s *OfflineStore) RemoveUser(ctx context.Context, luser, lserver string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM spool WHERE username=$1 AND server_host=$2`, luser, lserver)
	return err
}

func (s *OfflineStore) RemoveExpired(ctx context.Context, lserver string) (int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT username, xml, seq FROM spool WHERE server_host=$1 AND xml LIKE '%jabber:x:expire%'`, lserver)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	var expired []int64
	for rows.Next() {
		var user, xml string
		var seq int64
		if err := rows.Scan(&user, &xml, &seq); err != nil {
			rows.Close()
			return 0, err
		}
		if m, ok := decodeOffline(user, lserver, xml); ok && m.Expired(now) {
			expired = append(expired, seq)
		}
	}
	rows.Close()
	if len(expired) == 0 {
		return 0, rows.Err()
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM spool WHERE seq = ANY($1)`, pq.Array(expired))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
