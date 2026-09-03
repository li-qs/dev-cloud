package repo

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/user"
)

type User struct {
	user *ent.UserClient
}

func NewUser(db *ent.Client) *User {
	return &User{user: db.User}
}

func (r *User) GetByID(ctx context.Context, userID int) (*ent.User, error) {
	return r.user.Get(ctx, userID)
}

func (r *User) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.user.
		Query().
		Where(user.Username(username)).
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
