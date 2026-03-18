package cobalt

import (
	"encoding/json"
	"fmt"
)

type MetadataObject struct {
	Album       string `json:"album,omitempty"`
	Composer    string `json:"composer,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	AlbumArtist string `json:"album_artist,omitempty"`
	Track       string `json:"track,omitempty"`
	Date        string `json:"date,omitempty"`
	Sublanguage string `json:"sublanguage,omitempty"`
}

type OutputObject struct {
	Type      string          `json:"type"`
	Filename  string          `json:"filename"`
	Metadata  *MetadataObject `json:"metadata,omitempty"`
	Subtitles bool            `json:"subtitles"`
}

type AudioLocalProcessingObject struct {
	Copy      bool   `json:"copy"`
	Format    string `json:"format"`
	Bitrate   string `json:"bitrate"`
	Cover     bool   `json:"cover,omitempty"`
	CropCover bool   `json:"cropCover,omitempty"`
}

type PickerObject struct {
	Type  PickerType `json:"type"`
	Url   string     `json:"url"`
	Thumb string     `json:"thumb,omitempty"`
}

type ErrorObject struct {
	Code    string              `json:"code"`
	Context *ErrorContextObject `json:"context,omitempty"`
}

type ErrorContextObject struct {
	Service string  `json:"service,omitempty"`
	Limit   float64 `json:"limit,omitempty"`
}

type MainResponse struct {
	Status Status `json:"status"`

	// tunnel / redirect
	Url      string `json:"url,omitempty"`
	Filename string `json:"filename,omitempty"`

	// local-processing
	Type    LocalProcessingType         `json:"type,omitempty"`
	Service LocalProcessingService      `json:"service,omitempty"`
	Tunnel  []string                    `json:"tunnel,omitempty"`
	Output  *OutputObject               `json:"output,omitempty"`
	Audio   *AudioLocalProcessingObject `json:"audio,omitempty"`
	IsHLS   *bool                       `json:"isHLS,omitempty"`

	// picker
	PickerAudio   *string        `json:"-"`
	AudioFilename *string        `json:"audioFilename,omitempty"`
	Picker        []PickerObject `json:"picker,omitempty"`

	// error
	Error *ErrorObject `json:"error,omitempty"`
}

type MainResponseEnvelope struct {
	Status Status `json:"status"`
}

type TunnelOrRedirectResponse struct {
	Status   Status `json:"status"`
	Url      string `json:"url"`
	Filename string `json:"filename"`
}

type LocalProcessingResponse struct {
	Status  Status                      `json:"status"`
	Type    LocalProcessingType         `json:"type"`
	Service LocalProcessingService      `json:"service"`
	Tunnel  []string                    `json:"tunnel"`
	Output  OutputObject                `json:"output"`
	Audio   *AudioLocalProcessingObject `json:"audio,omitempty"`
	IsHLS   *bool                       `json:"isHLS,omitempty"`
}

type PickerResponse struct {
	Status        Status         `json:"status"`
	Audio         *string        `json:"audio,omitempty"`
	AudioFilename *string        `json:"audioFilename,omitempty"`
	Picker        []PickerObject `json:"picker"`
}

type ErrorResponse struct {
	Status Status      `json:"status"`
	Error  ErrorObject `json:"error"`
}

func ParseMainResponse(data []byte) (MainResponse, error) {
	var envelope MainResponseEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return MainResponse{}, fmt.Errorf("decode response envelope: %w", err)
	}

	switch envelope.Status {
	case StatusTunnel, StatusRedirect:
		var r TunnelOrRedirectResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return MainResponse{}, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return MainResponse{
			Status:   r.Status,
			Url:      r.Url,
			Filename: r.Filename,
		}, nil
	case StatusLocalProcessing:
		var r LocalProcessingResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return MainResponse{}, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return MainResponse{
			Status:  r.Status,
			Type:    r.Type,
			Service: r.Service,
			Tunnel:  r.Tunnel,
			Output:  &r.Output,
			Audio:   r.Audio,
			IsHLS:   r.IsHLS,
		}, nil
	case StatusPicker:
		var r PickerResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return MainResponse{}, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return MainResponse{
			Status:        r.Status,
			PickerAudio:   r.Audio,
			AudioFilename: r.AudioFilename,
			Picker:        r.Picker,
		}, nil
	case StatusError:
		var r ErrorResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return MainResponse{}, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return MainResponse{
			Status: r.Status,
			Error:  &r.Error,
		}, nil
	default:
		return MainResponse{}, fmt.Errorf("unsupported response status %q", envelope.Status)
	}
}

func PickerFilenameByType(objType PickerType, index int) string {
	switch objType {
	case PickerTypePhoto:
		return fmt.Sprintf("picker_photo_%d.jpg", index)
	case PickerTypeVideo:
		return fmt.Sprintf("picker_video_%d.mp4", index)
	default:
		return fmt.Sprintf("picker_file_%d", index)
	}
}
