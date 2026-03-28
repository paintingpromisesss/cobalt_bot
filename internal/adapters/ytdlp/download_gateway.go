package ytdlp

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/source"
)

type YtDLPDownloadGateway struct {
	client *Client
}

func NewDownloadGateway(client *Client) *YtDLPDownloadGateway {
	return &YtDLPDownloadGateway{client: client}
}

func (g *YtDLPDownloadGateway) IdentifyYoutubeURL(url string) (bool, media.YouTubeContentKind) {
	return g.client.IdentifyYoutubeURL(url)
}

func (g *YtDLPDownloadGateway) GetMetadata(ctx context.Context, url string) (*source.YtDLPMetadata, error) {
	meta, err := g.client.GetMetadata(ctx, url)
	if err != nil {
		return nil, err
	}

	result := &source.YtDLPMetadata{
		Title:              meta.Title,
		ThumbnailURL:       meta.Thumbnail,
		OriginalURL:        meta.OriginalURL,
		DurationSeconds:    meta.Duration,
		Formats:            make([]source.YtDLPFormat, 0, len(meta.Formats)),
		RequestedDownloads: make([]source.YtDLPRequestedDownload, 0, len(meta.RequestedDownloads)),
	}

	for _, format := range meta.Formats {
		if format.GetRoundedABR() == 0 && format.GetRoundedVBR() == 0 {
			continue
		}
		result.Formats = append(result.Formats, source.YtDLPFormat{
			FormatID:    format.FormatID,
			DisplayName: format.GetDisplayName(),
			FileSize:    format.FileSize,
			HasAudio:    format.IsAudio(),
			HasVideo:    format.IsVideo(),
		})
	}

	var bestAudioFormat *Format

	for _, download := range meta.RequestedDownloads {
		converted := source.YtDLPRequestedDownload{
			Formats: make([]source.YtDLPFormat, 0, len(download.RequestedFormats)),
		}
		bestAudioFormat = download.GetBestAudioFormat()
		for _, format := range download.RequestedFormats {
			converted.Formats = append(converted.Formats, source.YtDLPFormat{
				FormatID:    format.FormatID,
				DisplayName: format.GetDisplayName(),
				FileSize:    format.FileSize,
				HasAudio:    format.IsAudio(),
				HasVideo:    format.IsVideo(),
			})
		}
		result.RequestedDownloads = append(result.RequestedDownloads, converted)
	}

	if bestAudioFormat != nil {
		for _, format := range meta.Formats {
			if !format.IsVideo() || format.IsAudio() {
				continue
			}

			mergedFormat := MergeFormats(&format, bestAudioFormat)
			result.Formats = append(result.Formats, source.YtDLPFormat{
				FormatID:    mergedFormat.FormatID,
				DisplayName: mergedFormat.GetDisplayName(),
				FileSize:    mergedFormat.FileSize,
				HasAudio:    true,
				HasVideo:    true,
			})
		}
	}

	return result, nil
}
