package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/user"
)

const (
	UserStatusEnabled  = 1
	UserStatusDisabled = 0
)

type User struct {
	db *ent.Client
}

func NewUser(db *ent.Client) *User {
	return &User{db: db}
}

func (u *User) Create(
	ctx context.Context,
	username string,
	password string,
	nickname string,
) (*ent.User, error) {
	return u.db.User.
		Create().
		SetUsername(username).
		SetPassword(password).
		SetNickname(nickname).
		Save(ctx)
}

func (u *User) Delete(ctx context.Context, userID int) error {
	return u.db.User.
		UpdateOneID(userID).
		SetStatus(UserStatusDisabled).
		Exec(ctx)
}

func (u *User) Get(ctx context.Context, userID int) (*ent.User, error) {
	return u.db.User.Get(ctx, userID)
}

func (u *User) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return u.db.User.
		Query().
		Where(
			user.Username(username),
			user.StatusEQ(UserStatusEnabled),
		).
		Only(ctx)
}

func (u *User) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	_, err := u.db.User.
		UpdateOneID(userID).
		SetNickname(nickname).
		Save(ctx)
	return err
}

func (u *User) UpdatePassword(ctx context.Context, userID int, password string) error {
	_, err := u.db.User.
		UpdateOneID(userID).
		SetPassword(password).
		Save(ctx)
	return err
}
