package main

import (
	"FEAppPiA/internal/api"
	"FEAppPiA/internal/clipboardMng"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	ClipboardMng *clipboardMng.Manager
	apiServer    *api.Server
	mdnsServer   *zeroconf.Server
	authToken    string
	actualPort   int
}

// NewApp creates a new App application struct
func NewApp() *App {

	return &App{ClipboardMng: clipboardMng.NewManager(),
		authToken: generateRandomToken()}
}

func generateRandomToken() string {
	return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.ClipboardMng.StartMonitoring(ctx)

	a.startAPIServer()
	time.Sleep(200 * time.Millisecond)
	a.registerMDNSService()
	runtime.EventsEmit(a.ctx, "clipboard-updated", a.ClipboardMng)
}

func (a *App) startAPIServer() {
	// Создаём сервер с конфигурацией
	cfg := api.Config{
		Addr:         ":0", // Авто-выбор порта
		AuthToken:    a.authToken,
		ClipboardMng: a.ClipboardMng,
	}

	a.apiServer = api.NewServer(cfg)

	// Запускаем и получаем реальный порт
	port, err := a.apiServer.Start(a.ctx)
	if err != nil {
		fmt.Printf("❌ Ошибка запуска API: %v\n", err)
		return
	}
	a.actualPort = port
}

func (a *App) registerMDNSService() {
	hostname, _ := os.Hostname()
	//localIP := getLocalIP()

	// TXT-записи: токен и метаданные
	// ⚠️ zeroconf принимает []string в формате "key=value"
	txt := []string{
		fmt.Sprintf("token=%s", a.authToken),
		fmt.Sprintf("version=1.0"),
		fmt.Sprintf("port=%d", a.actualPort), // 👈 Передаём реальный порт!
	}

	// Регистрируем сервис
	// ⚠️ Порт должен быть реальным, а не ":0"
	server, err := zeroconf.Register(
		hostname,
		"_pcclip._tcp",
		"local.",
		a.actualPort,
		txt,
		nil,
	)

	if err != nil {
		fmt.Printf("❌ Ошибка регистрации mDNS: %v\n", err)
		return
	}

	a.mdnsServer = server
	fmt.Printf("🔍 mDNS: %s._pcclip._tcp.local.:%d\n", hostname, a.actualPort)
}

func (a *App) AddClipToHistory(content, Type string) {
	a.ClipboardMng.AddClipToHistory(content, Type)
}

func (a *App) GetClipboardHistory() []clipboardMng.Item {
	return a.ClipboardMng.GetClipboardHistory()
}

func (a *App) RestoreFromHistory(index int) error {
	return a.ClipboardMng.RestoreFromHistory(index)
}

func (a *App) GetPhoneAuthToken() string {
	return a.authToken
}

func (a *App) GetServerInfo() map[string]string {
	return map[string]string{
		"ip":    getLocalIP(),
		"port":  fmt.Sprintf("%d", a.actualPort),
		"token": a.authToken,
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}

func (a *App) Shutdown(ctx context.Context) {
	// 1. Останавливаем mDNS
	if a.mdnsServer != nil {
		a.mdnsServer.Shutdown()
	}

	// 2. Останавливаем HTTP сервер
	if a.apiServer != nil {
		if err := a.apiServer.Stop(ctx); err != nil {
			fmt.Printf("⚠️ Ошибка остановки API: %v\n", err)
		}
	}
}
