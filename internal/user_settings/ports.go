package user_settings

import "context"

type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (Settings, bool, error)
	UpsertByUserID(ctx context.Context, userID int64, settings Settings) error
}
