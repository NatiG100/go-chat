package drivers

import (
	"context"
	"errors"
	"time"

	"example.com/go-chat/internal/core"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Postgres struct {
	db *gorm.DB
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&core.Group{},
		&core.User{},
		&core.Message{},
		&core.GroupMember{},
	)
	if err!=nil{
		return nil, err
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}


func (p *Postgres) CreateUser(ctx context.Context, u *core.User, plainPassword string) error{
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil{
		return err
	}
	u.ID = uuid.New()
	u.PasswordHash = string(hash)
	u.CreatedAt = time.Now()
	return p.db.WithContext(ctx).Create(u).Error
}

func( p * Postgres) GetUserByEmail(ctx context.Context, email string) (*core .User, error){
	var u core.User
	err :=  p.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil{
		return nil, err
	}
return &u, nil
}

func (p * Postgres) GetUserByID(ctx context.Context, id uuid.UUID) (*core.User, error){
	var u core.User
	err := p.db.WithContext(ctx).First(&u, "id = ?",id).Error
	if err!= nil{
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) VerifyPassword(ctx context.Context, email, plain string) (*core.User, error){
	u, err := p.GetUserByEmail(ctx,email)
	if err!=nil{
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash),[]byte(plain)) !=nil{
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

func (p *Postgres) CreateGroup(ctx context.Context, g *core.Group) error{
	g.ID = uuid.New()
	g.CreatedAt = time.Now()
	return p.db.WithContext(ctx).Create(g).Error
}

type GroupMember struct{
	GroupID uuid.UUID `gorm:"type:uuid"`
	UserID uuid.UUID `gorm:"type:uuid"`
	JoinedAt time.Time
}

func (p *Postgres) AddGroupMember(ctx context.Context, groupId , userId uuid.UUID) error{
	m := core.GroupMember{
		GroupID:  groupId,
		UserID: userId,
		JoinedAt: time.Now(),
	}

	return p.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error
}
func (p *Postgres) MyGroups(ctx context.Context, userId uuid.UUID) ([]core.Group,error){
	var groups []core.Group

	err := p.db.WithContext(ctx).
		Joins("JOIN group_members gm ON gm.group_id = groups.id").
		Where("gm.user_id = ?", userId).Preload("Owner").
		Find(&groups).Error

	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (p *Postgres) ListGroupMembers(ctx context.Context, groupId uuid.UUID)([]core.User,error){
	var users []core.User

	err := p.db.WithContext(ctx).
	Joins("JOIN group_members gm ON gm.user_id=users.id").
	Where("gm.group_id = ?",groupId).Find(&users).Error
	return users,err
}


func (p *Postgres) SaveMessage(ctx context.Context, m *core.Message) error{
	m.ID = uuid.New()
	m.CreatedAt = time.Now()
	return p.db.WithContext(ctx).Create(m).Error
}

func (p *Postgres) GetPrivateHistory(ctx context.Context, a, b uuid.UUID, limit int)([]core.Message, error){
	var msgs []core.Message
	err:= p.db.WithContext(ctx).
	Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)", a,b,b,a).
	Order("created_at DESC").Limit(limit).Find(&msgs).Error
	return msgs,err
}

func (p *Postgres) GetGroupHistory(ctx context.Context, groupID uuid.UUID, limit int)([]core.Message,error){
	var msgs []core.Message
	err := p.db.WithContext(ctx).Where("group_id = ?",groupID).Order("created_at DESC").Limit(limit).Find(&msgs).Error

	return msgs,err
}

func NewRepositories(pg *Postgres) *Repositories{
	return &Repositories{pg}
}

type Repositories struct {*Postgres}

func (r *Repositories) UserRepo() core.UserRepository { return r }
func (r *Repositories) GroupRepo() core.GroupRepository { return r }
func (r *Repositories) MessageRepo() core.MessageRepository {return r }
