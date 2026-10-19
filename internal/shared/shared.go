package shared

import (
	"net/http"
	"os"

	"barbercentral-core/internal/shared/httpx"
)

type ErrorResponse = httpx.ErrorDetail

func RespondWithJSON(w http.ResponseWriter, status int, data any) {
	httpx.JSON(w, status, data)
}

func RespondWithError(w http.ResponseWriter, status int, message string, err error) {
	code := "ERROR"
	if err != nil && message == "" {
		message = err.Error()
	}
	httpx.Error(w, status, code, message)
}

func GetUploadsDir() string {
	dir := os.Getenv("UPLOADS_DIR")
	if dir != "" {
		return dir
	}
	if stat, err := os.Stat("/app"); err == nil && stat.IsDir() {
		return "/app/uploads"
	}
	return "../barbercentral-front/public/uploads"
}

