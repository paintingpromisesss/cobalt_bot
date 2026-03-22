package ytdlp

type YtDLPClient string

const (
	ClientDefault      YtDLPClient = "default"
	ClientWeb          YtDLPClient = "web"
	ClientWebEmbedded  YtDLPClient = "web_embedded"
	ClientWebSafari    YtDLPClient = "web_safari"
	ClientMWeb         YtDLPClient = "mweb"
	ClientWebMusic     YtDLPClient = "web_music"
	ClientWebCreator   YtDLPClient = "web_creator"
	ClientIOS          YtDLPClient = "ios"
	ClientAndroid      YtDLPClient = "android"
	ClientAndroidVR    YtDLPClient = "android_vr"
	ClientAndroidMusic YtDLPClient = "android_music"
	ClientTV           YtDLPClient = "tv"
	ClientTVDowngraded YtDLPClient = "tv_downgraded"
	ClientTVSimply     YtDLPClient = "tv_simply"
	ClientAll          YtDLPClient = "all"
)
