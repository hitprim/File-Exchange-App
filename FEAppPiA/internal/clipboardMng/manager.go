package clipboardMng

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.design/x/clipboard"
)

type Manager struct {
	history []Item
	maxClip int
	mu      sync.Mutex
	lastVal []byte
	ctx     context.Context
}

type Item struct {
	Content string    `json:"content"`
	Time    time.Time `json:"timestamp"`
	Type    string    `json:"type"`
}

func NewManager() *Manager {
	clipboard.Init()
	return &Manager{
		history: make([]Item, 0),
		maxClip: 10,
	}
}

func (manager *Manager) StartMonitoring(ctx context.Context) {
	manager.ctx = ctx
	manager.lastVal = clipboard.Read(clipboard.FmtText)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data := clipboard.Read(clipboard.FmtText)
				if len(data) > 0 && string(data) != string(manager.lastVal) {
					manager.lastVal = data
					manager.AddClipToHistory(string(data), "text")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (manager *Manager) AddClipToHistory(content, Type string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	newItem := Item{Content: content, Time: time.Now(), Type: Type}
	manager.history = append([]Item{newItem}, manager.history...)

	if len(manager.history) > manager.maxClip {
		manager.history = manager.history[:manager.maxClip]
	}
}

func (manager *Manager) GetClipboardHistory() []Item {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.history
}

func (manager *Manager) RestoreFromHistory(index int) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if index >= len(manager.history) || index < 0 {
		return fmt.Errorf("invalid index")
	}
	clipboard.Write(clipboard.FmtText, []byte(manager.history[index].Content))
	return nil
}

// Прямая запись в буфер (для приема с телефона)
func (manager *Manager) WriteToClipboard(content string) {
	if content == "" {
		return
	}
	clipboard.Write(clipboard.FmtText, []byte(content))
}
