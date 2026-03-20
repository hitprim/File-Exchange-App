package api

// ClipResponse — формат ответа для истории буфера
// Можно вынести, если нужно переиспользовать в других местах
type ClipResponse struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

// SendRequest — формат запроса от телефона
type SendRequest struct {
	Text string `json:"text"`
}

// SendResponse — ответ на отправку
type SendResponse struct {
	Status string `json:"status"`
}
