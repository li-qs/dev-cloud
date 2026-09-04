package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/refreshtoken"
	"time"
)

type RefreshToken struct {
	token *ent.RefreshTokenClient
}

func NewToken(db *ent.Client) *RefreshToken {
	return &RefreshToken{token: db.RefreshToken}
}

func (r *RefreshToken) Get(ctx context.Context, tokenHash string) (*ent.RefreshToken, error) {
	return r.token.
		Query().
		Where(refreshtoken.TokenHash(tokenHash)).
		Only(ctx)
}

func (r *RefreshToken) Create(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.token.
		Create().
		SetUserID(userID).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *RefreshToken) Delete(ctx context.Context, id int) error {
	return r.token.DeleteOneID(id).Exec(ctx)
}

func (r *RefreshToken) DeleteByUserID(ctx context.Context, userID int) error {
	_, err := r.token.
		Delete().
		Where(refreshtoken.UserID(userID)).
		Exec(ctx)
	return err
}
