package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/refreshtoken"
	"time"
)

type RefreshToken struct {
	db *ent.Client
}

func NewToken(db *ent.Client) *RefreshToken {
	return &RefreshToken{db: db}
}

func (r *RefreshToken) Get(ctx context.Context, tokenHash string) (*ent.RefreshToken, error) {
	return r.db.RefreshToken.
		Query().
		Where(refreshtoken.TokenHash(tokenHash)).
		Only(ctx)
}

func (r *RefreshToken) Create(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.RefreshToken.
		Create().
		SetUserID(userID).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *RefreshToken) Delete(ctx context.Context, id int) error {
	return r.db.RefreshToken.DeleteOneID(id).Exec(ctx)
}

func (r *RefreshToken) DeleteByUserID(ctx context.Context, userID int) error {
	_, err := r.db.RefreshToken.
		Delete().
		Where(refreshtoken.UserID(userID)).
		Exec(ctx)
	return err
}
