package download

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/user_settings"
)

type Service struct {
	settings        SettingsService
	cobalt          cobalt.CobaltClient
	queue           Queue
	tempDir         string
	maxFileBytes    int64
	downloadTimeout time.Duration
}

func NewService(settings SettingsService, cobaltClient cobalt.CobaltClient, requestQueue Queue, tempDir string, maxFileBytes int64, downloadTimeout time.Duration) (*Service, error) {
	if settings == nil {
		return nil, fmt.Errorf("settings service is nil")
	}
	if cobaltClient == nil {
		return nil, fmt.Errorf("cobalt client is nil")
	}
	if requestQueue == nil {
		return nil, fmt.Errorf("request queue is nil")
	}
	if strings.TrimSpace(tempDir) == "" {
		return nil, fmt.Errorf("temp dir is empty")
	}
	if maxFileBytes <= 0 {
		return nil, fmt.Errorf("max file bytes must be positive, got %d", maxFileBytes)
	}
	if downloadTimeout <= 0 {
		return nil, fmt.Errorf("download timeout must be positive, got %s", downloadTimeout)
	}

	return &Service{
		settings:        settings,
		cobalt:          cobaltClient,
		queue:           requestQueue,
		tempDir:         tempDir,
		maxFileBytes:    maxFileBytes,
		downloadTimeout: downloadTimeout,
	}, nil
}

func (s *Service) Execute(ctx context.Context, req Request) (Result, error) {
	if err := req.Validate(); err != nil {
		return Result{}, err
	}

	var result Result
	err := s.queue.Run(req.UserID, func() error {
		settings, err := s.settings.GetByUserID(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("load user settings: %w", err)
		}

		processResp, err := s.cobalt.Process(ctx, mapProcessRequest(req.SourceURL, settings))
		if err != nil {
			return err
		}

		switch processResp.Status {
		case "tunnel", "redirect", "local-processing":
			file, err := s.downloadByURL(ctx, processResp.URL, processResp.Filename)
			if err != nil {
				return err
			}
			result = Result{
				Status: processResp.Status,
				File:   &file,
			}
			return nil

		case "picker":
			options := parsePickerOptions(processResp.Picker)
			if len(options) == 0 {
				return fmt.Errorf("cobalt returned empty picker options")
			}

			result = Result{
				Status:        processResp.Status,
				PickerOptions: options,
			}

			if req.PickerIndex == nil {
				return nil
			}

			selected, ok := findPickerOption(options, *req.PickerIndex)
			if !ok {
				return fmt.Errorf("invalid picker option index: %d", *req.PickerIndex)
			}

			file, err := s.downloadByURL(ctx, selected.URL, selected.Filename)
			if err != nil {
				return err
			}
			result.File = &file
			return nil

		case "error":
			if processResp.Error != nil {
				return &cobalt.CobaltAPIError{Code: processResp.Error.Code, Context: processResp.Error.Context}
			}
			return fmt.Errorf("cobalt returned status=error")

		default:
			return fmt.Errorf("unsupported cobalt response status: %q", processResp.Status)
		}
	})
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

func mapProcessRequest(sourceURL string, settings user_settings.Settings) cobalt.CobaltProcessRequest {
	return cobalt.CobaltProcessRequest{
		URL:                   sourceURL,
		AudioBitrate:          settings.AudioBitrate,
		AudioFormat:           settings.AudioFormat,
		DownloadMode:          settings.DownloadMode,
		FilenameStyle:         settings.FilenameStyle,
		VideoQuality:          settings.VideoQuality,
		SubtitleLang:          settings.SubtitleLang,
		YoutubeVideoCodec:     settings.YoutubeVideoCodec,
		YoutubeVideoContainer: settings.YoutubeVideoContainer,
		YoutubeBetterAudio:    settings.YoutubeBetterAudio,
	}
}

func (s *Service) downloadByURL(ctx context.Context, downloadURL, filenameHint string) (cobalt.CobaltDownloadedFile, error) {
	downloadURL = strings.TrimSpace(downloadURL)
	if downloadURL == "" {
		return cobalt.CobaltDownloadedFile{}, fmt.Errorf("cobalt response has empty download url")
	}

	downloadCtx, cancel := context.WithTimeout(ctx, s.downloadTimeout)
	defer cancel()

	return s.cobalt.Download(downloadCtx, cobalt.CobaltDownloadRequest{
		URL:          downloadURL,
		TempDir:      s.tempDir,
		FilenameHint: filenameHint,
		MaxFileBytes: s.maxFileBytes,
	})
}

func parsePickerOptions(raw []map[string]any) []PickerOption {
	options := make([]PickerOption, 0, len(raw))
	for _, item := range raw {
		url := readString(item, "url")
		if strings.TrimSpace(url) == "" {
			continue
		}

		label := readString(item, "label")
		if label == "" {
			label = readString(item, "name")
		}
		if label == "" {
			label = readString(item, "type")
		}
		if label == "" {
			label = readString(item, "quality")
		}
		if label == "" {
			label = "Вариант"
		}

		options = append(options, PickerOption{
			Index:    len(options),
			URL:      url,
			Label:    label,
			Filename: readString(item, "filename"),
		})
	}

	return options
}

func findPickerOption(options []PickerOption, index int) (PickerOption, bool) {
	for _, option := range options {
		if option.Index == index {
			return option, true
		}
	}
	return PickerOption{}, false
}

func readString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	value, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
