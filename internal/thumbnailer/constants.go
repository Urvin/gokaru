package thumbnailer

import "gopkg.in/gographics/imagick.v3/imagick"

const (
	// CAST_RESIZE_TENSILE Растяжение
	CAST_RESIZE_TENSILE = 2

	// CAST_RESIZE_PRECISE Вписывает в нужный размер по максимуму
	CAST_RESIZE_PRECISE = 4

	// CAST_RESIZE_INVERSE Вписывает в нужный размер по минимуму
	CAST_RESIZE_INVERSE = 8

	// CAST_TRIM Обрезает поля изображения
	CAST_TRIM = 16

	// CAST_EXTENT Устанавливает канву изрбражения нужного размера
	CAST_EXTENT = 32

	// CAST_OPAQUE_BACKGROUND Установить непрозрачный задний фон
	CAST_OPAQUE_BACKGROUND = 64

	// CAST_TRANSPARENT_BACKGROUND Установить прозрачный задний фон
	CAST_TRANSPARENT_BACKGROUND = 128

	// CAST_TRIM_PADDING Добавляет 10 пикселей вокруг изображения при обрезке полей
	CAST_TRIM_PADDING = 256

	FORMAT_JPG  = "jpg"
	FORMAT_PNG  = "png"
	FORMAT_WEBP = "webp"
	FORMAT_GIF  = "gif"

	RESIZE_FILTER = imagick.FILTER_SINC
)

var FORMATS_ALLOWED = []string{FORMAT_JPG, FORMAT_PNG, FORMAT_WEBP, FORMAT_GIF}
var FORMATS_OPAQUE = []string{FORMAT_JPG, FORMAT_PNG}
var FORMATS_ANIMATED = []string{FORMAT_WEBP, FORMAT_GIF}
