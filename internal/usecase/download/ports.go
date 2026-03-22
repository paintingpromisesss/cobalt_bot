package download

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/source"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
	usecasesettings "github.com/paintingpromisesss/cobalt_bot/internal/usecase/settings"
)

type SettingsService interface {
	GetOrCreateUserSettings(ctx context.Context, userID int64) (user.Settings, error)
}

type URLPolicy interface {
	Validate(raw string) (string, bool)
}

type CobaltGateway interface {
	GetContent(ctx context.Context, url string, settings user.Settings) (source.CobaltContent, error)
}

type YTDLPGateway interface {
	IdentifyYoutubeURL(url string) (bool, media.YouTubeContentKind)
	GetMetadata(ctx context.Context, url string) (*source.YtDLPMetadata, error)
}

var _ SettingsService = (*usecasesettings.Service)(nil)
