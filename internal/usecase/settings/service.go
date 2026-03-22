package settings

import (
	"context"
	"errors"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
)

type Service struct {
	repo                    Repository
	errUserSettingsNotFound error
}

func NewService(repo Repository, errUserSettingsNotFound error) *Service {
	return &Service{
		repo:                    repo,
		errUserSettingsNotFound: errUserSettingsNotFound,
	}
}

func (s *Service) GetOrCreateUserSettings(ctx context.Context, userID int64) (user.Settings, error) {
	settings, err := s.repo.GetUserSettings(ctx, userID)
	if err == nil {
		return settings, nil
	}

	if s.errUserSettingsNotFound == nil || !errors.Is(err, s.errUserSettingsNotFound) {
		return user.Settings{}, err
	}

	settings = user.DefaultSettings()
	settings.UserID = userID

	if err := s.repo.UpsertUserSettings(ctx, settings); err != nil {
		return user.Settings{}, err
	}

	return settings, nil
}

func (s *Service) UpsertUserSettings(ctx context.Context, settings user.Settings) error {
	return s.repo.UpsertUserSettings(ctx, settings)
}
