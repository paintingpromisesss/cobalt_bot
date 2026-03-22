package start

import (
	"context"
	"strings"

	usecasesettings "github.com/paintingpromisesss/cobalt_bot/internal/usecase/settings"
)

type Service struct {
	settings          SettingsService
	availableServices []string
}

func NewService(settings SettingsService, availableServices []string) *Service {
	return &Service{
		settings:          settings,
		availableServices: availableServices,
	}
}

func (s *Service) Handle(ctx context.Context, input Input) (Result, error) {
	if _, err := s.settings.GetOrCreateUserSettings(ctx, input.UserID); err != nil {
		return Result{}, err
	}

	return Result{
		Message: "Бот запущен. Просто отправьте ссылку на контент, который хотите скачать. Доступные сервисы: " + strings.Join(s.availableServices, ", "),
	}, nil
}

var _ SettingsService = (*usecasesettings.Service)(nil)
