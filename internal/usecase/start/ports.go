package start

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
)

type SettingsService interface {
	GetOrCreateUserSettings(ctx context.Context, userID int64) (user.Settings, error)
}
