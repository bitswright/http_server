package httpcore

const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusInternalServerError = 500
)

const (
	StatusTextOK                  = "OK"
	StatusTextBadRequest          = "Bad Request"
	StatusTextNotFound            = "Not Found"
	StatusTextMethodNotAllowed    = "Method Not Allowed"
	StatusTextInternalServerError = "Internal Server Error"
)

func getStatusText(status int) string {
	statusTextMap := map[int]string{
		StatusOK:                  StatusTextOK,
		StatusBadRequest:          StatusTextBadRequest,
		StatusNotFound:            StatusTextNotFound,
		StatusMethodNotAllowed:    StatusTextMethodNotAllowed,
		StatusInternalServerError: StatusTextInternalServerError,
	}
	if text, ok := statusTextMap[status]; ok {
		return text
	}
	return "Unknown"
}
