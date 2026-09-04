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
	user *ent.UserClient
}

func NewUser(db *ent.Client) *User {
	return &User{user: db.User}
}

func (u *User) Create(
	ctx context.Context,
	username string,
	password string,
	nickname string,
) (*ent.User, error) {
	return u.user.
		Create().
		SetUsername(username).
		SetPassword(password).
		SetNickname(nickname).
		Save(ctx)
}

func (u *User) Delete(ctx context.Context, userID int) error {
	return u.user.
		Update().
		SetStatus(UserStatusDisabled).
		Exec(ctx)
}

func (r *User) Get(ctx context.Context, userID int) (*ent.User, error) {
	return r.user.Get(ctx, userID)
}

func (r *User) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.user.
		Query().
		Where(
			user.Username(username),
			user.StatusEQ(UserStatusEnabled),
		).
		Only(ctx)
}

func (r *User) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	_, err := r.user.
		UpdateOneID(userID).
		SetNickname(nickname).
		Save(ctx)
	return err
}

func (r *User) UpdatePassword(ctx context.Context, userID int, password string) error {
	_, err := r.user.
		UpdateOneID(userID).
		SetPassword(password).
		Save(ctx)
	return err
}
