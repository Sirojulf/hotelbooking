package repository

import (
	"fmt"
	"sync"
	"time"

	"hotelbooking/internal/config"
	"hotelbooking/internal/models"

	json "github.com/goccy/go-json"
)

// TTL cache sederhana untuk data yang jarang berubah (room, room_type).
// Aman dipakai concurrently. TTL pendek (60s) — kalau ada update via API,
// cache otomatis expired dalam waktu wajar.

const cacheTTL = 60 * time.Second

type roomCacheEntry struct {
	room   *models.Room
	expiry time.Time
}

type roomTypeCacheEntry struct {
	roomType *models.RoomType
	expiry   time.Time
}

var (
	roomCacheMu sync.RWMutex
	roomCache   = make(map[string]roomCacheEntry)

	roomTypeCacheMu sync.RWMutex
	roomTypeCache   = make(map[string]roomTypeCacheEntry)
)

func getCachedRoom(id string) *models.Room {
	roomCacheMu.RLock()
	defer roomCacheMu.RUnlock()
	if e, ok := roomCache[id]; ok && time.Now().Before(e.expiry) {
		return e.room
	}
	return nil
}

func setCachedRoom(id string, r *models.Room) {
	roomCacheMu.Lock()
	roomCache[id] = roomCacheEntry{room: r, expiry: time.Now().Add(cacheTTL)}
	roomCacheMu.Unlock()
}

func invalidateRoom(id string) {
	roomCacheMu.Lock()
	delete(roomCache, id)
	roomCacheMu.Unlock()
}

func getCachedRoomType(id string) *models.RoomType {
	roomTypeCacheMu.RLock()
	defer roomTypeCacheMu.RUnlock()
	if e, ok := roomTypeCache[id]; ok && time.Now().Before(e.expiry) {
		return e.roomType
	}
	return nil
}

func setCachedRoomType(id string, rt *models.RoomType) {
	roomTypeCacheMu.Lock()
	roomTypeCache[id] = roomTypeCacheEntry{roomType: rt, expiry: time.Now().Add(cacheTTL)}
	roomTypeCacheMu.Unlock()
}

func invalidateRoomType(id string) {
	roomTypeCacheMu.Lock()
	delete(roomTypeCache, id)
	roomTypeCacheMu.Unlock()
}

// WarmRoomCache pre-load semua room & room_type ke cache saat startup.
// Eliminasi cache-miss di request pertama yang biasanya bikin P99 melonjak.
// Dipanggil dari main.go setelah ConnectSupabase().
func WarmRoomCache() {
	if config.SupabaseClient == nil {
		return
	}

	if resp, _, err := config.SupabaseClient.From("rooms").Select("*", "", false).Execute(); err == nil {
		var rooms []models.Room
		if jerr := json.Unmarshal(resp, &rooms); jerr == nil {
			for i := range rooms {
				setCachedRoom(rooms[i].ID.String(), &rooms[i])
			}
			fmt.Printf("Pre-warmed %d rooms ke cache\n", len(rooms))
		}
	}

	if resp, _, err := config.SupabaseClient.From("room_types").Select("*", "", false).Execute(); err == nil {
		var rts []models.RoomType
		if jerr := json.Unmarshal(resp, &rts); jerr == nil {
			for i := range rts {
				setCachedRoomType(rts[i].ID.String(), &rts[i])
			}
			fmt.Printf("Pre-warmed %d room_types ke cache\n", len(rts))
		}
	}
}