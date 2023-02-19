package main

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type memorySticker struct {
	sticker  *Sticker
	lastSeen time.Time
}

type MemSticker struct {
	mu              sync.Mutex
	userStickersMap map[int64]*memorySticker
}

func newMemSticker() *MemSticker {
	return &MemSticker{userStickersMap: make(map[int64]*memorySticker)}
}

func (m *MemSticker) addSticker(uid int64, tgSticker *tgbotapi.Sticker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userStickersMap[uid] = &memorySticker{
		sticker: &Sticker{
			uniqueID: tgSticker.FileUniqueID,
			id:       tgSticker.FileID,
			userID:   uid,
		},
		lastSeen: time.Now(),
	}
}

func (m *MemSticker) updateSticker(uid int64, tags string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stick, ok := m.userStickersMap[uid]
	if !ok {
		return ErrRecordNotFound
	}
	stick.sticker.tags = tags
	return nil
}

func (m *MemSticker) backgroundCleanup() {
	for {
		time.Sleep(time.Minute)
		m.mu.Lock()
		for user, sticker := range m.userStickersMap {
			if time.Since(sticker.lastSeen) > 5*time.Minute {
				delete(m.userStickersMap, user)
			}
		}
		m.mu.Unlock()
	}
}
