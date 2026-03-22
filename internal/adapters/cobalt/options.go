package cobalt

type AudioBitrate string
type AudioFormat string
type DownloadMode string
type FilenameStyle string
type VideoQuality string
type LocalProcessing string
type SubtitleLanguage string
type YoutubeVideoCodec string
type YoutubeVideoContainer string

type Status string
type LocalProcessingType string
type LocalProcessingService string
type PickerType string

const (
	Bitrate320 AudioBitrate = "320"
	Bitrate256 AudioBitrate = "256"
	Bitrate128 AudioBitrate = "128"
	Bitrate96  AudioBitrate = "96"
	Bitrate64  AudioBitrate = "64"
	Bitrate8   AudioBitrate = "8"

	FormatBest AudioFormat = "best"
	FormatMP3  AudioFormat = "mp3"
	FormatOGG  AudioFormat = "ogg"
	FormatWAV  AudioFormat = "wav"
	FormatOPUS AudioFormat = "opus"

	ModeAuto  DownloadMode = "auto"
	ModeAudio DownloadMode = "audio"
	ModeMute  DownloadMode = "mute"

	StyleClassic FilenameStyle = "classic"
	StylePretty  FilenameStyle = "pretty"
	StyleBasic   FilenameStyle = "basic"
	StyleNerdy   FilenameStyle = "nerdy"

	QualityMax  VideoQuality = "max"
	Quality4320 VideoQuality = "4320"
	Quality2160 VideoQuality = "2160"
	Quality1440 VideoQuality = "1440"
	Quality1080 VideoQuality = "1080"
	Quality720  VideoQuality = "720"
	Quality480  VideoQuality = "480"
	Quality360  VideoQuality = "360"
	Quality240  VideoQuality = "240"
	Quality144  VideoQuality = "144"

	ProcessingDisabled  LocalProcessing = "disabled"
	ProcessingPreferred LocalProcessing = "preferred"
	ProcessingForced    LocalProcessing = "forced"

	YoutubeCodecH264 YoutubeVideoCodec = "h264"
	YoutubeCodecAV1  YoutubeVideoCodec = "av1"
	YoutubeCodecVP9  YoutubeVideoCodec = "vp9"

	YoutubeContainerAuto YoutubeVideoContainer = "auto"
	YoutubeContainerMP4  YoutubeVideoContainer = "mp4"
	YoutubeContainerWEBM YoutubeVideoContainer = "webm"
	YoutubeContainerMKV  YoutubeVideoContainer = "mkv"

	StatusTunnel          Status = "tunnel"
	StatusLocalProcessing Status = "local-processing"
	StatusRedirect        Status = "redirect"
	StatusPicker          Status = "picker"
	StatusError           Status = "error"

	LocalProcessingMerge LocalProcessingType = "merge"
	LocalProcessingMute  LocalProcessingType = "mute"
	LocalProcessingAudio LocalProcessingType = "audio"
	LocalProcessingGif   LocalProcessingType = "gif"
	LocalProcessingRemux LocalProcessingType = "remux"

	LocalProcessingServiceBilibili    LocalProcessingService = "bilibili"
	LocalProcessingServiceBluesky     LocalProcessingService = "bluesky"
	LocalProcessingServiceDailymotion LocalProcessingService = "dailymotion"
	LocalProcessingServiceInstagram   LocalProcessingService = "instagram"
	LocalProcessingServiceFacebook    LocalProcessingService = "facebook"
	LocalProcessingServiceLoom        LocalProcessingService = "loom"
	LocalProcessingServiceNewgrounds  LocalProcessingService = "newgrounds"
	LocalProcessingServiceOkRu        LocalProcessingService = "ok"
	LocalProcessingServicePinterest   LocalProcessingService = "pinterest"
	LocalProcessingServiceReddit      LocalProcessingService = "reddit"
	LocalProcessingServiceRutube      LocalProcessingService = "rutube"
	LocalProcessingServiceSnapchat    LocalProcessingService = "snapchat"
	LocalProcessingServiceSoundcloud  LocalProcessingService = "soundcloud"
	LocalProcessingServiceStreamable  LocalProcessingService = "streamable"
	LocalProcessingServiceTiktok      LocalProcessingService = "tiktok"
	LocalProcessingServiceTumblr      LocalProcessingService = "tumblr"
	LocalProcessingServiceTwitch      LocalProcessingService = "twitch"
	LocalProcessingServiceTwitter     LocalProcessingService = "twitter"
	LocalProcessingServiceVimeo       LocalProcessingService = "vimeo"
	LocalProcessingServiceVk          LocalProcessingService = "vk"
	LocalProcessingServiceXiaohongshu LocalProcessingService = "xiaohongshu"
	LocalProcessingServiceYoutube     LocalProcessingService = "youtube"

	PickerTypePhoto PickerType = "photo"
	PickerTypeVideo PickerType = "video"
	PickerTypeGif   PickerType = "gif"
)
