package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

// IdempotencyEntry menyimpan snapshot respons HTTP dan hash payload untuk suatu key.
type IdempotencyEntry struct {
	RequestHash string
	StatusCode  int
	Body        []byte
	Headers     map[string]string
	CreatedAt   time.Time
}

// IdempotencyStore adalah interface untuk penyimpanan idempotency entry.
type IdempotencyStore interface {
	Get(key string) (*IdempotencyEntry, bool)
	Set(key string, entry *IdempotencyEntry, ttl time.Duration)
}

// MemoryIdempotencyStore menyimpan entri idempotency di dalam memori secara thread-safe.
type MemoryIdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*IdempotencyEntry
}

// NewMemoryIdempotencyStore membuat instansiasi MemoryIdempotencyStore baru.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	store := &MemoryIdempotencyStore{
		entries: make(map[string]*IdempotencyEntry),
	}
	// Background cleanup goroutine sederhana untuk membersihkan entry expired setiap 5 menit
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			store.cleanup()
		}
	}()
	return store
}

func (s *MemoryIdempotencyStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.entries {
		// Default eviction jika sudah lewat 1 jam
		if now.Sub(v.CreatedAt) > 1*time.Hour {
			delete(s.entries, k)
		}
	}
}

// Get mengambil entry dari memori store.
func (s *MemoryIdempotencyStore) Get(key string) (*IdempotencyEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, found := s.entries[key]
	return entry, found
}

// Set menyimpan entry ke memori store.
func (s *MemoryIdempotencyStore) Set(key string, entry *IdempotencyEntry, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = entry
}

// IdempotencyConfig menyimpan konfigurasi middleware Idempotency.
type IdempotencyConfig struct {
	// HeaderName adalah nama header idempotensi (default: "Idempotency-Key")
	HeaderName string
	// TTL adalah masa berlaku idempotency entry (default: 30 menit)
	TTL time.Duration
	// Store adalah backend penyimpanan (default: MemoryIdempotencyStore)
	Store IdempotencyStore
}

var defaultMemoryStore = NewMemoryIdempotencyStore()

// IdempotencyGuard mengembalikan middleware Fiber v3 yang melindungi endpoint dari duplicate replay.
// Jika header Idempotency-Key tidak dikirim oleh client, request akan dilanjutkan normal tanpa blocking (opt-in).
func IdempotencyGuard(config ...IdempotencyConfig) fiber.Handler {
	cfg := IdempotencyConfig{
		HeaderName: "Idempotency-Key",
		TTL:        30 * time.Minute,
		Store:      defaultMemoryStore,
	}
	if len(config) > 0 {
		if config[0].HeaderName != "" {
			cfg.HeaderName = config[0].HeaderName
		}
		if config[0].TTL > 0 {
			cfg.TTL = config[0].TTL
		}
		if config[0].Store != nil {
			cfg.Store = config[0].Store
		}
	}

	return func(c fiber.Ctx) error {
		key := c.Get(cfg.HeaderName)
		// Jika client tidak menyertakan header Idempotency-Key, pass-through langsung (backward compatible)
		if key == "" {
			return c.Next()
		}

		// Identifikasi request: Path, Method, UserID (jika terotentikasi), dan Request Body
		userID := ""
		if uid := c.Locals("user_id"); uid != nil {
			userID = fmt.Sprintf("%v", uid)
		}

		bodyBytes := c.Body()
		hasher := sha256.New()
		hasher.Write([]byte(c.Method()))
		hasher.Write([]byte(c.Path()))
		hasher.Write([]byte(userID))
		hasher.Write(bodyBytes)
		currentHash := hex.EncodeToString(hasher.Sum(nil))

		// Cek apakah key sudah pernah digunakan sebelumnya
		if entry, exists := cfg.Store.Get(key); exists {
			// Periksa apakah payload sama
			if entry.RequestHash != currentHash {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"success": false,
					"status":  "error",
					"message": "Idempotency key telah digunakan dengan payload request yang berbeda",
				})
			}

			// Payload identik: replay respons sebelumnya tanpa memanggil handler lagi
			for hKey, hVal := range entry.Headers {
				c.Set(hKey, hVal)
			}
			c.Set("X-Cache-Lookup", "HIT - Idempotent Replay")
			return c.Status(entry.StatusCode).Send(entry.Body)
		}

		// Jalankan handler utama
		if err := c.Next(); err != nil {
			return err
		}

		// Hanya simpan respons sukses / valid (biasanya status < 500)
		statusCode := c.Response().StatusCode()
		if statusCode < 500 {
			respBody := make([]byte, len(c.Response().Body()))
			copy(respBody, c.Response().Body())

			cachedHeaders := make(map[string]string)
			// Simpan content-type jika ada
			if ct := c.Response().Header.ContentType(); len(ct) > 0 {
				cachedHeaders[fiber.HeaderContentType] = string(ct)
			}

			newEntry := &IdempotencyEntry{
				RequestHash: currentHash,
				StatusCode:  statusCode,
				Body:        respBody,
				Headers:     cachedHeaders,
				CreatedAt:   time.Now(),
			}
			cfg.Store.Set(key, newEntry, cfg.TTL)
		}

		return nil
	}
}
