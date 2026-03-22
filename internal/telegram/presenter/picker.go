package presenter

import (
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	tele "gopkg.in/telebot.v4"
)

const (
	NeutralIndicator         = "⬜"
	SelectedIndicator        = "✅"
	UnselectedIndicator      = "❌"
	DownloadIndicator        = "⬇️"
	CobaltPickerButtonUnique = "cobalt_picker_button"
	YtDLPPickerButtonUnique  = "ytdlp_picker_button"
)

func BuildCobaltPickerMessage(sessionID string, pickerView *picker.CobaltView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	total := len(pickerView.Options)
	rows := make([]tele.Row, 0, total+3)
	selected := 0

	for i, option := range pickerView.Options {
		indicator := NeutralIndicator
		if option.Selected {
			indicator = SelectedIndicator
			selected++
		}
		payload := EncodeCobaltPickerCallbackData(picker.CobaltActionToggle, sessionID, i)
		rows = append(rows, markup.Row(markup.Data(indicator+" "+option.Label, CobaltPickerButtonUnique, payload)))
	}

	if selected > 0 {
		rows = append(rows, markup.Row(
			markup.Data(UnselectedIndicator+" Очистить выбор", CobaltPickerButtonUnique, EncodeCobaltPickerCallbackData(picker.CobaltActionClearAll, sessionID, -1)),
			markup.Data(DownloadIndicator+" Скачать", CobaltPickerButtonUnique, EncodeCobaltPickerCallbackData(picker.CobaltActionDownload, sessionID, -1)),
		))
	} else {
		rows = append(rows, markup.Row(
			markup.Data(SelectedIndicator+" Выбрать все", CobaltPickerButtonUnique, EncodeCobaltPickerCallbackData(picker.CobaltActionSelectAll, sessionID, -1)),
		))
	}

	rows = append(rows, markup.Row(
		markup.Data(UnselectedIndicator+" Отменить", CobaltPickerButtonUnique, EncodeCobaltPickerCallbackData(picker.CobaltActionCancel, sessionID, -1)),
	))

	markup.Inline(rows...)

	message := fmt.Sprintf("Найдено файлов: %d. Выбрано: %d.\n Отметьте нужные и нажмите «Скачать».", total, selected)
	return markup, message
}

func BuildYtDLPPickerMessage(sessionID string, pickerView *picker.YtDLPView) (*tele.ReplyMarkup, string) {
	if pickerView.ActiveTab == picker.YtDLPTabNone {
		return buildYtDLPTabsMessage(sessionID, pickerView)
	}
	return buildYtDLPOptionsMessage(sessionID, pickerView)
}

func BuildYtDLPConfirmationMessage(sessionID string, option picker.YtDLPOption) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, 2)

	downloadPayload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionDownload, sessionID, picker.YtDLPTabNone, -1)
	backPayload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionConfirmBack, sessionID, picker.YtDLPTabNone, -1)

	rows = append(rows, markup.Row(markup.Data("Скачать", YtDLPPickerButtonUnique, downloadPayload)))
	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, backPayload)))

	markup.Inline(rows...)

	message := fmt.Sprintf("Выбранный формат: %s\n Размер: %s\n Скачать?", option.DisplayName, formatFileSize(option.FileSize))
	return markup, message
}

func buildYtDLPTabsMessage(sessionID string, pickerView *picker.YtDLPView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(pickerView.Tabs)+1)
	for _, tab := range pickerView.Tabs {
		payload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionTab, sessionID, tab, -1)
		rows = append(rows, markup.Row(markup.Data(getYtDLPTabLabel(tab), YtDLPPickerButtonUnique, payload)))
	}

	cancelPayload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionCancel, sessionID, picker.YtDLPTabNone, -1)
	rows = append(rows, markup.Row(markup.Data("Отменить", YtDLPPickerButtonUnique, cancelPayload)))

	markup.Inline(rows...)
	message := fmt.Sprintf("Скачиваемый контент: %s \n Выберите опцию скачивания:", pickerView.ContentName)
	return markup, message
}

func buildYtDLPOptionsMessage(sessionID string, pickerView *picker.YtDLPView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(pickerView.Options)+1)

	for i, option := range pickerView.Options {
		payload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionChoose, sessionID, pickerView.ActiveTab, i)
		rows = append(rows, markup.Row(markup.Data(option.DisplayName, YtDLPPickerButtonUnique, payload)))
	}

	backPayload := EncodeYtDLPPickerCallbackData(picker.YtDLPActionBack, sessionID, picker.YtDLPTabNone, -1)
	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, backPayload)))

	markup.Inline(rows...)
	message := fmt.Sprintf("Выберите формат скачивания для: %s\n (тип: %s)", pickerView.ContentName, pickerView.ActiveTab)
	return markup, message
}

func getYtDLPTabLabel(tab picker.YtDLPTab) string {
	switch tab {
	case picker.YtDLPTabAudioOnly:
		return "Аудио (только звук)"
	case picker.YtDLPTabVideoOnly:
		return "Видео (только картинка)"
	case picker.YtDLPTabAudioVideo:
		return "Аудио + Видео (звук и картинка в одном файле)"
	default:
		return "Неизвестно"
	}
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return formatFloatTrimmed(float64(size)/GB) + " GB"
	case size >= MB:
		return formatFloatTrimmed(float64(size)/MB) + " MB"
	case size >= KB:
		return formatFloatTrimmed(float64(size)/KB) + " KB"
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func formatFloatTrimmed(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
