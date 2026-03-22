package download

import (
	"context"
	"fmt"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	domainpicker "github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/source"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
)

type Service struct {
	settings  SettingsService
	urlPolicy URLPolicy
	cobalt    CobaltGateway
	ytdlp     YTDLPGateway
}

func NewService(
	settings SettingsService,
	urlPolicy URLPolicy,
	cobalt CobaltGateway,
	ytdlp YTDLPGateway,
) *Service {
	return &Service{
		settings:  settings,
		urlPolicy: urlPolicy,
		cobalt:    cobalt,
		ytdlp:     ytdlp,
	}
}

func (s *Service) Handle(ctx context.Context, input Input) (Result, error) {
	normalizedURL, ok := s.urlPolicy.Validate(input.URL)
	if !ok {
		return InvalidURLResult{}, nil
	}

	settings, err := s.settings.GetOrCreateUserSettings(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return s.resolveDownload(ctx, normalizedURL, settings)
}

func (s *Service) resolveDownload(ctx context.Context, url string, settings user.Settings) (Result, error) {
	isYoutubeURL, contentType := s.ytdlp.IdentifyYoutubeURL(url)
	if !isYoutubeURL {
		return s.resolveCobalt(ctx, url, settings)
	}

	return s.resolveYtDLP(ctx, url, contentType)
}

func (s *Service) resolveCobalt(ctx context.Context, url string, settings user.Settings) (Result, error) {
	resp, err := s.cobalt.GetContent(ctx, url, settings)
	if err != nil {
		return nil, err
	}

	switch resp.Status {
	case source.CobaltStatusRedirect, source.CobaltStatusTunnel:
		return CobaltDirectResult{
			URL: url,
			File: media.RemoteFile{
				URL:      resp.FileURL,
				Filename: resp.FileName,
			},
		}, nil
	case source.CobaltStatusPicker:
		options := make([]domainpicker.CobaltOption, len(resp.Options))
		for i, opt := range resp.Options {
			options[i] = domainpicker.CobaltOption{
				Label:    opt.Label,
				URL:      opt.URL,
				Filename: opt.Filename,
			}
		}
		return CobaltPickerResult{
			URL: url,
			Data: domainpicker.CobaltInitData{
				Options: options,
			},
		}, nil
	case source.CobaltStatusError:
		return nil, cobaltErrorToErr(resp.Error)
	default:
		return nil, fmt.Errorf("unsupported cobalt status: %q", resp.Status)
	}
}

func (s *Service) resolveYtDLP(ctx context.Context, url string, contentType media.YouTubeContentKind) (Result, error) {
	meta, err := s.ytdlp.GetMetadata(ctx, url)
	if err != nil {
		return nil, err
	}

	switch contentType {
	case media.YouTubeVideo:
		return YtDLPPickerResult{URL: url, Data: *buildYtDLPPickerInitData(meta)}, nil
	case media.YouTubeMusic:
		option, err := buildBestAudioOption(meta)
		if err != nil {
			return nil, err
		}
		return YtDLPDirectResult{URL: url, Option: *option}, nil
	case media.YouTubeShorts:
		option, err := buildBestAudioVideoOption(meta)
		if err != nil {
			return nil, err
		}
		return YtDLPDirectResult{URL: url, Option: *option}, nil
	default:
		return nil, fmt.Errorf("unsupported youtube content type: %q", contentType)
	}
}

func buildBestAudioOption(meta *source.YtDLPMetadata) (*domainpicker.YtDLPOption, error) {
	var bestAudioFormat *source.YtDLPFormat
	for _, requestedDownload := range meta.RequestedDownloads {
		bestAudioFormat = requestedDownload.GetBestAudioFormat()
		break
	}
	if bestAudioFormat == nil {
		return nil, fmt.Errorf("не найден подходящий аудио формат для скачивания")
	}

	option := domainpicker.YtDLPOption{
		DisplayName:  bestAudioFormat.DisplayName,
		ThumbnailURL: meta.ThumbnailURL,
		ContentURL:   meta.OriginalURL,
		FormatID:     bestAudioFormat.FormatID,
		FileSize:     bestAudioFormat.FileSize,
		Duration:     time.Duration(meta.DurationSeconds) * time.Second,
		Format:       media.DownloadFormat{HasAudio: true},
	}

	return &option, nil
}

func buildBestAudioVideoOption(meta *source.YtDLPMetadata) (*domainpicker.YtDLPOption, error) {
	var bestVideoFormat *source.YtDLPFormat
	var bestAudioFormat *source.YtDLPFormat
	for _, requestedDownload := range meta.RequestedDownloads {
		bestVideoFormat = requestedDownload.GetBestVideoFormat()
		bestAudioFormat = requestedDownload.GetBestAudioFormat()
		break
	}
	if bestVideoFormat == nil {
		return nil, fmt.Errorf("не найден подходящий видео формат для скачивания")
	}
	if bestAudioFormat == nil {
		return nil, fmt.Errorf("не найден подходящий аудио формат для скачивания")
	}

	option := domainpicker.YtDLPOption{
		DisplayName:  bestVideoFormat.DisplayName,
		ThumbnailURL: meta.ThumbnailURL,
		ContentURL:   meta.OriginalURL,
		FormatID:     bestVideoFormat.FormatID + "+" + bestAudioFormat.FormatID,
		FileSize:     bestVideoFormat.FileSize + bestAudioFormat.FileSize,
		Duration:     time.Duration(meta.DurationSeconds) * time.Second,
		Format:       media.DownloadFormat{HasAudio: true, HasVideo: true},
	}

	return &option, nil
}

func buildYtDLPPickerInitData(meta *source.YtDLPMetadata) *domainpicker.YtDLPInitData {
	optionsByTab := make(map[domainpicker.YtDLPTab][]domainpicker.YtDLPOption)

	for _, format := range meta.Formats {
		tab := detectTabForFormat(format)
		if tab == "" {
			continue
		}

		option := domainpicker.YtDLPOption{
			DisplayName:  format.DisplayName,
			ThumbnailURL: meta.ThumbnailURL,
			ContentURL:   meta.OriginalURL,
			FormatID:     format.FormatID,
			FileSize:     format.FileSize,
			Duration:     time.Duration(meta.DurationSeconds) * time.Second,
			Format: media.DownloadFormat{
				HasAudio: format.HasAudio,
				HasVideo: format.HasVideo,
			},
		}

		optionsByTab[tab] = append(optionsByTab[tab], option)
	}

	if len(meta.RequestedDownloads) > 0 {
		bestAudioFormat := meta.RequestedDownloads[0].GetBestAudioFormat()
		if bestAudioFormat != nil {
			for _, format := range meta.Formats {
				if detectTabForFormat(format) != domainpicker.YtDLPTabVideoOnly {
					continue
				}

				option := domainpicker.YtDLPOption{
					DisplayName:  format.DisplayName,
					ContentURL:   meta.OriginalURL,
					FormatID:     format.FormatID + "+" + bestAudioFormat.FormatID,
					ThumbnailURL: meta.ThumbnailURL,
					FileSize:     bestAudioFormat.FileSize + format.FileSize,
					Duration:     time.Duration(meta.DurationSeconds) * time.Second,
					Format: media.DownloadFormat{
						HasAudio: true,
						HasVideo: true,
					},
				}
				optionsByTab[domainpicker.YtDLPTabAudioVideo] = append(optionsByTab[domainpicker.YtDLPTabAudioVideo], option)
			}
		}
	}

	return &domainpicker.YtDLPInitData{
		ContentName:  meta.Title,
		OptionsByTab: optionsByTab,
	}
}

func detectTabForFormat(format source.YtDLPFormat) domainpicker.YtDLPTab {
	switch {
	case format.HasVideo && format.HasAudio:
		return domainpicker.YtDLPTabAudioVideo
	case format.HasVideo && !format.HasAudio:
		return domainpicker.YtDLPTabVideoOnly
	case !format.HasVideo && format.HasAudio:
		return domainpicker.YtDLPTabAudioOnly
	default:
		return ""
	}
}

func cobaltErrorToErr(errObj *source.CobaltError) error {
	if errObj == nil {
		return fmt.Errorf("cobalt returned unspecified error")
	}

	switch {
	case errObj.Service != "":
		return fmt.Errorf("cobalt error: %s (service=%s)", errObj.Code, errObj.Service)
	case errObj.Limit != 0:
		return fmt.Errorf("cobalt error: %s (limit=%v)", errObj.Code, errObj.Limit)
	default:
		return fmt.Errorf("cobalt error: %s", errObj.Code)
	}
}
