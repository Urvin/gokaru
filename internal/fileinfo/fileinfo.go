package fileinfo

import (
	"mime"
	"net/http"
	"strings"
)

func MimeByData(data []byte) string {
	return http.DetectContentType(data)
}

func MimeByExtension(extension string) string {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return mime.TypeByExtension(extension)
}
