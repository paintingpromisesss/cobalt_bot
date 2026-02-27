package download

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/user_settings"
)

type SettingsService interface {
	GetByUserID(ctx context.Context, userID int64) (user_settings.Settings, error)
}

type Queue interface {
	Run(userID int64, fn func() error) error
}
