package drivers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"example.com/go-chat/internal/core"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Postgres{pool: p}, nil
}

func (p *Postgres) Close(ctx context.Context) { p.pool.Close() }

// -- User operations
func (p *Postgres) CreateUser(ctx context.Context, u *core.User, plainPassword string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.ID = uuid.New()
	u.PasswordHash = string(h)
	q := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ($1,$2,$3,$4,$5)`
	_, err = p.pool.Exec(ctx, q, u.ID, u.Username, u.Email, u.PasswordHash, time.Now())
	return err
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*core.User, error) {
	q := `SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1`
	row := p.pool.QueryRow(ctx, q, email)
	var u core.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id uuid.UUID) (*core.User, error) {
	q := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1`
	row := p.pool.QueryRow(ctx, q, id)
	var u core.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) VerifyPassword(ctx context.Context, email, plain string) (*core.User, error) {
	u, err := p.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// -- Groups
func (p *Postgres) CreateGroup(ctx context.Context, g *core.Group) error {
	g.ID = uuid.New()
	q := `INSERT INTO groups (id, name, owner_id, created_at) VALUES ($1,$2,$3,$4)`
	_, err := p.pool.Exec(ctx, q, g.ID, g.Name, g.OwnerID, time.Now())
	return err
}

func (p *Postgres) AddGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	q := `INSERT INTO group_members (group_id, user_id, joined_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`
	_, err := p.pool.Exec(ctx, q, groupID, userID, time.Now())
	return err
}

func (p *Postgres) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]core.User, error) {
	q := `SELECT u.id, u.username, u.email, u.password_hash, u.created_at FROM users u JOIN group_members gm ON u.id = gm.user_id WHERE gm.group_id=$1`
	rows, err := p.pool.Query(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.User
	for rows.Next() {
		var u core.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

// -- Messages
func (p *Postgres) SaveMessage(ctx context.Context, m *core.Message) error {
	m.ID = uuid.New()
	q := `INSERT INTO messages (id, sender_id, recipient_id, group_id, content, created_at) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := p.pool.Exec(ctx, q, m.ID, m.SenderID, m.RecipientID, m.GroupID, m.Content, time.Now())
	return err
}

func (p *Postgres) GetPrivateHistory(ctx context.Context, a, b uuid.UUID, limit int) ([]core.Message, error) {
	q := `SELECT id, sender_id, recipient_id, group_id, content, created_at FROM messages WHERE (sender_id=$1 AND recipient_id=$2) OR (sender_id=$2 AND recipient_id=$1) ORDER BY created_at DESC LIMIT $3`
	rows, err := p.pool.Query(ctx, q, a, b, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Message
	for rows.Next() {
		var m core.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.GroupID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (p *Postgres) GetGroupHistory(ctx context.Context, groupID uuid.UUID, limit int) ([]core.Message, error) {
	q := `SELECT id, sender_id, recipient_id, group_id, content, created_at FROM messages WHERE group_id=$1 ORDER BY created_at DESC LIMIT $2`
	rows, err := p.pool.Query(ctx, q, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Message
	for rows.Next() {
		var m core.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.GroupID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// convenience
func NewRepositories(pg *Postgres) *Repositories {
	return &Repositories{pg}
}

// implement core.Repositories by embedding Postgres
type Repositories struct{ *Postgres }

func (r *Repositories) UserRepo() core.UserRepository     { return r }
func (r *Repositories) GroupRepo() core.GroupRepository   { return r }
func (r *Repositories) MessageRepo() core.MessageRepository { return r }
