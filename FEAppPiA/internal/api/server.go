package api

import (
	"FEAppPiA/internal/clipboardMng"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Config — конфигурация сервера
type Config struct {
	Addr         string // Адрес для прослушки, например ":0" для авто-порта
	AuthToken    string // Токен для аутентификации
	ClipboardMng *clipboardMng.Manager
}

// Server — HTTP-сервер для мобильного клиента
type Server struct {
	cfg    Config
	server *http.Server
	ctx    context.Context
}

// NewServer создаёт новый экземпляр сервера
func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg}
}

// Start запускает сервер в горутине
// Возвращает реальный порт (если в addr был ":0") и ошибку
func (s *Server) Start(ctx context.Context) (int, error) {
	s.ctx = ctx

	mux := http.NewServeMux()

	// Регистрируем хендлеры
	mux.HandleFunc("/api/clipboard", s.handleGetClipboard)
	mux.HandleFunc("/api/clipboard/send", s.handleReceiveFromPhone)

	// Применяем middleware
	handler := s.authMiddleware(mux)

	s.server = &http.Server{
		Addr:    s.cfg.Addr,
		Handler: handler,
		// Таймауты для безопасности
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Слушаем на всех интерфейсах
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}

	// Получаем реальный порт
	actualPort := ln.Addr().(*net.TCPAddr).Port

	// Запускаем сервер в горутине
	go func() {
		fmt.Printf("📱 HTTP-сервер запущен на порту %d\n", actualPort)
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ Ошибка сервера: %v\n", err)
		}
	}()

	return actualPort, nil
}

// Stop корректно останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ==================== Middleware ====================

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем токен в query param или заголовке
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.Header.Get("X-Auth-Token")
		}
		if token != s.cfg.AuthToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Разрешаем CORS для мобильного клиента
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Token")

		// Обработка preflight запросов
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ==================== Хендлеры ====================

// GET /api/clipboard — получить историю буфера
func (s *Server) handleGetClipboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := s.cfg.ClipboardMng.GetClipboardHistory()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// POST /api/clipboard/send — отправить текст с телефона на ПК
func (s *Server) handleReceiveFromPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text != "" {
		// Копируем в системный буфер ПК
		s.cfg.ClipboardMng.WriteToClipboard(req.Text)
		// Добавляем в историю
		s.cfg.ClipboardMng.AddClipToHistory(req.Text, "from_phone")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
