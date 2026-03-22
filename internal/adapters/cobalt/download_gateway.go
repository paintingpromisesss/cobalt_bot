package cobalt

import (
	"context"
	"fmt"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/source"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
)

type CobaltDownloadGateway struct {
	client *Client
}

func NewDownloadGateway(client *Client) *CobaltDownloadGateway {
	return &CobaltDownloadGateway{client: client}
}

func (g *CobaltDownloadGateway) GetContent(ctx context.Context, url string, settings user.Settings) (source.CobaltContent, error) {
	resp, err := g.client.GetContentURL(ctx, NewRequest(url, settings))
	if err != nil {
		return source.CobaltContent{}, err
	}

	result := source.CobaltContent{
		Status:   source.CobaltStatus(resp.Status),
		FileURL:  resp.Url,
		FileName: resp.Filename,
	}

	if resp.Error != nil {
		result.Error = &source.CobaltError{
			Code: resp.Error.Code,
		}
		if resp.Error.Context != nil {
			result.Error.Service = resp.Error.Context.Service
			result.Error.Limit = resp.Error.Context.Limit
		}
	}

	options := make([]source.CobaltOption, 0, len(resp.Picker)+1)
	for i, obj := range resp.Picker {
		options = append(options, source.CobaltOption{
			Label:    fmt.Sprintf("%s #%d", strings.ToUpper(string(obj.Type)), i+1),
			URL:      obj.Url,
			Filename: PickerFilenameByType(obj.Type, i+1),
		})
	}

	if resp.PickerAudio != nil && resp.AudioFilename != nil {
		options = append(options, source.CobaltOption{
			Label:    "Аудио",
			URL:      *resp.PickerAudio,
			Filename: *resp.AudioFilename,
		})
	}

	result.Options = options
	return result, nil
}
