package media

type DownloadFormat struct {
	HasAudio bool
	HasVideo bool
}

func (f DownloadFormat) IsAudio() bool {
	return f.HasAudio
}

func (f DownloadFormat) IsVideo() bool {
	return f.HasVideo
}
