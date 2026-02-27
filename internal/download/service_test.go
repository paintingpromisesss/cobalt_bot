package download

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/queue"
	"github.com/paintingpromisesss/cobalt_bot/internal/user_settings"
)

type stubSettingsService struct {
	out        user_settings.Settings
	err        error
	lastUserID int64
}

func (s *stubSettingsService) GetByUserID(_ context.Context, userID int64) (user_settings.Settings, error) {
	s.lastUserID = userID
	if s.err != nil {
		return user_settings.Settings{}, s.err
	}
	return s.out, nil
}

type stubQueue struct {
	err        error
	lastUserID int64
}

func (q *stubQueue) Run(userID int64, fn func() error) error {
	q.lastUserID = userID
	if q.err != nil {
		return q.err
	}
	return fn()
}

type stubCobaltClient struct {
	processResp      cobalt.CobaltProcessResponse
	processErr       error
	downloadResp     cobalt.CobaltDownloadedFile
	downloadErr      error
	lastProcessReq   cobalt.CobaltProcessRequest
	lastDownloadReq  cobalt.CobaltDownloadRequest
	downloadCallCnt  int
	processCallCount int
}

func (c *stubCobaltClient) Process(_ context.Context, req cobalt.CobaltProcessRequest) (cobalt.CobaltProcessResponse, error) {
	c.processCallCount++
	c.lastProcessReq = req
	if c.processErr != nil {
		return cobalt.CobaltProcessResponse{}, c.processErr
	}
	return c.processResp, nil
}

func (c *stubCobaltClient) Download(_ context.Context, req cobalt.CobaltDownloadRequest) (cobalt.CobaltDownloadedFile, error) {
	c.downloadCallCnt++
	c.lastDownloadReq = req
	if c.downloadErr != nil {
		return cobalt.CobaltDownloadedFile{}, c.downloadErr
	}
	return c.downloadResp, nil
}

func TestNewServiceValidation(t *testing.T) {
	settings := &stubSettingsService{}
	client := &stubCobaltClient{}
	requestQueue := &stubQueue{}

	_, err := NewService(nil, client, requestQueue, "./tmp", 1, time.Second)
	if err == nil {
		t.Fatalf("expected nil settings validation error")
	}

	_, err = NewService(settings, nil, requestQueue, "./tmp", 1, time.Second)
	if err == nil {
		t.Fatalf("expected nil cobalt validation error")
	}

	_, err = NewService(settings, client, nil, "./tmp", 1, time.Second)
	if err == nil {
		t.Fatalf("expected nil queue validation error")
	}

	_, err = NewService(settings, client, requestQueue, " ", 1, time.Second)
	if err == nil {
		t.Fatalf("expected empty temp dir validation error")
	}

	_, err = NewService(settings, client, requestQueue, "./tmp", 0, time.Second)
	if err == nil {
		t.Fatalf("expected max file bytes validation error")
	}

	_, err = NewService(settings, client, requestQueue, "./tmp", 1, 0)
	if err == nil {
		t.Fatalf("expected timeout validation error")
	}
}

func TestExecuteTunnelDownloadsWithMappedSettings(t *testing.T) {
	settings := &stubSettingsService{
		out: user_settings.Settings{
			VideoQuality:          "720",
			DownloadMode:          "audio",
			AudioFormat:           "opus",
			AudioBitrate:          "96",
			FilenameStyle:         "pretty",
			YoutubeVideoCodec:     "vp9",
			YoutubeVideoContainer: "webm",
			YoutubeBetterAudio:    true,
			SubtitleLang:          "ru",
		},
	}
	client := &stubCobaltClient{
		processResp: cobalt.CobaltProcessResponse{
			Status:   "tunnel",
			URL:      "http://cobalt/tunnel?id=1",
			Filename: "video.mp4",
		},
		downloadResp: cobalt.CobaltDownloadedFile{
			Path:      "/tmp/file.bin",
			Filename:  "video.mp4",
			SizeBytes: 123,
		},
	}
	requestQueue := &stubQueue{}

	svc, err := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	if err != nil {
		t.Fatalf("init service: %v", err)
	}

	result, err := svc.Execute(context.Background(), Request{
		UserID:    42,
		SourceURL: "https://example.com/video",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if settings.lastUserID != 42 || requestQueue.lastUserID != 42 {
		t.Fatalf("expected user id propagation, got settings=%d queue=%d", settings.lastUserID, requestQueue.lastUserID)
	}
	if client.lastProcessReq.URL != "https://example.com/video" {
		t.Fatalf("unexpected process url: %q", client.lastProcessReq.URL)
	}
	if client.lastProcessReq.VideoQuality != "720" || client.lastProcessReq.AudioFormat != "opus" {
		t.Fatalf("settings were not mapped into process request: %+v", client.lastProcessReq)
	}
	if client.lastDownloadReq.URL != "http://cobalt/tunnel?id=1" {
		t.Fatalf("unexpected download url: %q", client.lastDownloadReq.URL)
	}
	if client.lastDownloadReq.TempDir != "./tmp" || client.lastDownloadReq.MaxFileBytes != 1024 {
		t.Fatalf("unexpected download request params: %+v", client.lastDownloadReq)
	}
	if result.Status != "tunnel" || result.File == nil {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestExecutePickerWithoutSelectionReturnsOptions(t *testing.T) {
	settings := &stubSettingsService{out: user_settings.DefaultSettings()}
	client := &stubCobaltClient{
		processResp: cobalt.CobaltProcessResponse{
			Status: "picker",
			Picker: []map[string]any{
				{"url": "https://a", "label": "720p"},
				{"url": "https://b", "type": "audio"},
			},
		},
	}
	requestQueue := &stubQueue{}

	svc, _ := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	result, err := svc.Execute(context.Background(), Request{
		UserID:    1,
		SourceURL: "https://example.com/video",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if result.Status != "picker" {
		t.Fatalf("unexpected status: %q", result.Status)
	}
	if len(result.PickerOptions) != 2 {
		t.Fatalf("expected 2 picker options, got %d", len(result.PickerOptions))
	}
	if client.downloadCallCnt != 0 {
		t.Fatalf("download must not be called without picker selection")
	}
}

func TestExecutePickerWithSelectionDownloads(t *testing.T) {
	settings := &stubSettingsService{out: user_settings.DefaultSettings()}
	client := &stubCobaltClient{
		processResp: cobalt.CobaltProcessResponse{
			Status: "picker",
			Picker: []map[string]any{
				{"url": "https://a", "label": "A"},
				{"url": "https://b", "label": "B", "filename": "b.mp4"},
			},
		},
		downloadResp: cobalt.CobaltDownloadedFile{
			Path:     "/tmp/b",
			Filename: "b.mp4",
		},
	}
	requestQueue := &stubQueue{}

	svc, _ := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	choice := 1
	result, err := svc.Execute(context.Background(), Request{
		UserID:      1,
		SourceURL:   "https://example.com/video",
		PickerIndex: &choice,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if client.downloadCallCnt != 1 {
		t.Fatalf("expected one download call, got %d", client.downloadCallCnt)
	}
	if client.lastDownloadReq.URL != "https://b" {
		t.Fatalf("expected picker option url to be downloaded, got %q", client.lastDownloadReq.URL)
	}
	if result.File == nil || result.File.Filename != "b.mp4" {
		t.Fatalf("unexpected result file: %+v", result.File)
	}
}

func TestExecutePickerInvalidSelection(t *testing.T) {
	settings := &stubSettingsService{out: user_settings.DefaultSettings()}
	client := &stubCobaltClient{
		processResp: cobalt.CobaltProcessResponse{
			Status: "picker",
			Picker: []map[string]any{
				{"url": "https://a"},
			},
		},
	}
	requestQueue := &stubQueue{}

	svc, _ := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	choice := 9
	_, err := svc.Execute(context.Background(), Request{
		UserID:      1,
		SourceURL:   "https://example.com/video",
		PickerIndex: &choice,
	})
	if err == nil || !strings.Contains(err.Error(), "invalid picker option index") {
		t.Fatalf("expected invalid picker index error, got %v", err)
	}
}

func TestExecutePropagatesQueueError(t *testing.T) {
	settings := &stubSettingsService{out: user_settings.DefaultSettings()}
	client := &stubCobaltClient{}
	requestQueue := &stubQueue{err: queue.ErrUserJobInProgress}

	svc, _ := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	_, err := svc.Execute(context.Background(), Request{
		UserID:    1,
		SourceURL: "https://example.com/video",
	})
	if !errors.Is(err, queue.ErrUserJobInProgress) {
		t.Fatalf("expected queue error, got %v", err)
	}
}

func TestExecutePropagatesSettingsError(t *testing.T) {
	settings := &stubSettingsService{err: errors.New("db fail")}
	client := &stubCobaltClient{}
	requestQueue := &stubQueue{}

	svc, _ := NewService(settings, client, requestQueue, "./tmp", 1024, 5*time.Second)
	_, err := svc.Execute(context.Background(), Request{
		UserID:    1,
		SourceURL: "https://example.com/video",
	})
	if err == nil || !strings.Contains(err.Error(), "load user settings") {
		t.Fatalf("expected wrapped settings error, got %v", err)
	}
}
