package cache

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CacheConfig represents cache middleware configuration
type CacheConfig struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c *fiber.Ctx) bool

	// Expiration is the time in which the cached entry will expire
	Expiration time.Duration

	// CacheControl enables the cache-control header
	CacheControl bool

	// Methods defines the HTTP methods that will be cached
	Methods []string

	// KeyGenerator defines a function to generate the cache key
	KeyGenerator func(c *fiber.Ctx) string

	// CacheHeader defines the header name for cache status
	CacheHeader string

	// ExcludeHeaders defines headers to exclude from caching
	ExcludeHeaders []string

	// CacheService is the Redis cache service instance
	CacheService *CacheService
}

// ConfigDefault is the default config
var ConfigDefault = CacheConfig{
	Next:         nil,
	Expiration:   1 * time.Minute,
	CacheControl: false,
	Methods:      []string{fiber.MethodGet, fiber.MethodHead},
	KeyGenerator: defaultKeyGenerator,
	CacheHeader:  "X-Cache",
	ExcludeHeaders: []string{
		"Authorization",
		"Cookie",
		"Set-Cookie",
	},
}

// Redis cache middleware for Fiber
func NewMiddleware(config ...CacheConfig) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided(Cập nhật cấu hình nếu có)
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.KeyGenerator == nil {
			cfg.KeyGenerator = defaultKeyGenerator
		}
		if cfg.CacheHeader == "" {
			cfg.CacheHeader = "X-Cache"
		}
		if cfg.Methods == nil {
			cfg.Methods = []string{fiber.MethodGet, fiber.MethodHead}
		}
	}

	// Initialize cache service if not (khởi tạo cache service nếu chưa có)
	if cfg.CacheService == nil {
		cacheService, err := NewCacheService()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize cache service: %v", err))
		}
		cfg.CacheService = cacheService
	}

	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns (Kiểm tra nếu Next trả về true để bỏ qua middleware)
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Check if method should be cached
		method := c.Method()
		shouldCache := false
		for _, m := range cfg.Methods {
			if method == m {
				shouldCache = true
				break
			}
		}

		if !shouldCache {
			return c.Next()
		}

		// Generate cache key
		key := cfg.KeyGenerator(c)

		// Try to get from cache
		cachedBody, err := cfg.CacheService.Get(key)
		if err == nil {
			// Cache hit
			c.Set(cfg.CacheHeader, "HIT")

			// Parse cached response
			cachedResp := &CachedResponse{}
			if err := cfg.CacheService.GetStruct(key+"_meta", cachedResp); err == nil {
				// Set headers from cache
				for k, v := range cachedResp.Headers {
					if !isExcludedHeader(k, cfg.ExcludeHeaders) {
						c.Set(k, v)
					}
				}
				c.Status(cachedResp.StatusCode)
			}

			if cfg.CacheControl {
				c.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(cfg.Expiration.Seconds())))
			}

			return c.SendString(cachedBody)
		}

		// Cache miss - execute handler
		c.Set(cfg.CacheHeader, "MISS")

		// Continue with the request
		if err := c.Next(); err != nil {
			return err
		}

		// Cache the response if status is 200 and method is cacheable
		if c.Response().StatusCode() == 200 && shouldCache {
			// Get response body
			body := string(c.Response().Body())

			// Create cached response metadata
			cachedResp := &CachedResponse{
				StatusCode: c.Response().StatusCode(),
				Headers:    make(map[string]string),
			}

			// Store headers (excluding sensitive ones)
			c.Response().Header.VisitAll(func(key, value []byte) {
				headerName := string(key)
				if !isExcludedHeader(headerName, cfg.ExcludeHeaders) {
					cachedResp.Headers[headerName] = string(value)
				}
			})

			// Cache both body and metadata
			go func() {
				cfg.CacheService.Set(key, body, cfg.Expiration)
				cfg.CacheService.Set(key+"_meta", cachedResp, cfg.Expiration)
			}()

			if cfg.CacheControl {
				c.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(cfg.Expiration.Seconds())))
			}
		}

		return nil
	}
}

// CachedResponse represents cached response metadata
type CachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

// Default key generator
func defaultKeyGenerator(c *fiber.Ctx) string {
	return generateCacheKey(c.Method(), c.OriginalURL(), string(c.Request().URI().QueryString()))
}

// Generate cache key from request
func generateCacheKey(method, url, query string) string {
	key := fmt.Sprintf("%s:%s", method, url)
	if query != "" {
		key += "?" + query
	}

	// Create MD5 hash of the key to keep it short
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("cache:http:%x", hash)
}

// Check if header should be excluded from caching
func isExcludedHeader(header string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if strings.EqualFold(header, excluded) {
			return true
		}
	}
	return false
}

// Session middleware using Redis()(Mục tiêu: Quản lý session của người dùng (lưu trữ và lấy thông tin session từ Redis).)
func SessionMiddleware(config ...SessionConfig) fiber.Handler {
	cfg := DefaultSessionConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheService == nil {
		cacheService, err := NewCacheService()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize cache service: %v", err))
		}
		cfg.CacheService = cacheService
	}

	return func(c *fiber.Ctx) error {
		// Skip if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get session ID from cookie or header
		sessionID := c.Cookies(cfg.CookieName)
		if sessionID == "" && cfg.HeaderName != "" {
			sessionID = c.Get(cfg.HeaderName)
		}

		// Create new session if not exists
		if sessionID == "" {
			sessionID = generateSessionID()
		}

		// Create session key
		sessionKey := fmt.Sprintf("session:%s", sessionID)

		// Get session data
		sessionData := make(map[string]interface{})
		if err := cfg.CacheService.GetStruct(sessionKey, &sessionData); err != nil {
			// No session data found, create empty session
			sessionData = make(map[string]interface{})
		}

		// Create session object
		session := &Session{
			ID:           sessionID,
			Data:         sessionData,
			cacheService: cfg.CacheService,
			key:          sessionKey,
			expiration:   cfg.Expiration,
		}

		// Store session in context
		c.Locals(cfg.ContextKey, session)

		// Set cookie if new session
		if c.Cookies(cfg.CookieName) == "" {
			c.Cookie(&fiber.Cookie{
				Name:     cfg.CookieName,
				Value:    sessionID,
				Expires:  time.Now().Add(cfg.Expiration),
				HTTPOnly: cfg.HTTPOnly,
				Secure:   cfg.Secure,
				SameSite: cfg.SameSite,
			})
		}

		return c.Next()
	}
}

// SessionConfig represents session middleware configuration
type SessionConfig struct {
	Next         func(c *fiber.Ctx) bool
	CookieName   string
	HeaderName   string
	ContextKey   string
	Expiration   time.Duration
	HTTPOnly     bool
	Secure       bool
	SameSite     string
	CacheService *CacheService
}

// DefaultSessionConfig is the default session config
var DefaultSessionConfig = SessionConfig{
	Next:       nil,
	CookieName: "session_id",
	HeaderName: "X-Session-ID",
	ContextKey: "session",
	Expiration: 24 * time.Hour,
	HTTPOnly:   true,
	Secure:     false,
	SameSite:   "Lax",
}

// Session represents a user session
type Session struct {
	ID           string
	Data         map[string]interface{}
	cacheService *CacheService
	key          string
	expiration   time.Duration
}

// Get gets a value from session
func (s *Session) Get(key string) interface{} {
	return s.Data[key]
}

// Set sets a value in session
func (s *Session) Set(key string, value interface{}) {
	s.Data[key] = value
}

// Delete deletes a value from session
func (s *Session) Delete(key string) {
	delete(s.Data, key)
}

// Save saves session to Redis
func (s *Session) Save() error {
	return s.cacheService.Set(s.key, s.Data, s.expiration)
}

// Destroy destroys the session
func (s *Session) Destroy() error {
	s.Data = make(map[string]interface{})
	return s.cacheService.Delete(s.key)
}

// Rate Limiting using Redis()(Mục tiêu: Giới hạn số lượng yêu cầu mà một người dùng hoặc một địa chỉ IP có thể thực hiện trong một khoảng thời gian nhất định (rate limiting).)
func RateLimitMiddleware(config ...RateLimitConfig) fiber.Handler {
	cfg := DefaultRateLimitConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheService == nil {
		cacheService, err := NewCacheService()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize cache service: %v", err))
		}
		cfg.CacheService = cacheService
	}

	return func(c *fiber.Ctx) error {
		// Skip if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get identifier (IP, User ID, etc.)
		identifier := cfg.KeyGenerator(c)
		key := fmt.Sprintf("ratelimit:%s", identifier)

		// Get current count
		count, err := cfg.CacheService.Get(key)
		if err != nil {
			// First request
			count = "0"
		}

		currentCount, _ := strconv.ParseInt(count, 10, 64)

		// Check if limit exceeded
		if currentCount >= cfg.Max {
			c.Set("X-RateLimit-Limit", strconv.FormatInt(cfg.Max, 10))
			c.Set("X-RateLimit-Remaining", "0")

			// Get TTL for reset time
			ttl, _ := cfg.CacheService.GetTTL(key)
			c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			return c.Status(cfg.StatusCode).JSON(fiber.Map{
				"error":   true,
				"message": cfg.Message,
			})
		}

		// Increment counter
		newCount, err := cfg.CacheService.IncrementBy(key, 1)
		if err != nil {
			newCount = 1
			cfg.CacheService.Set(key, newCount, cfg.Duration)
		} else if newCount == 1 {
			// Set expiration for first request
			cfg.CacheService.SetExpire(key, cfg.Duration)
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.FormatInt(cfg.Max, 10))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(cfg.Max-newCount, 10))

		// Get TTL for reset time
		ttl, _ := cfg.CacheService.GetTTL(key)
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

		return c.Next()
	}
}

// RateLimitConfig represents rate limit configuration
type RateLimitConfig struct {
	Next         func(c *fiber.Ctx) bool
	Max          int64
	Duration     time.Duration
	KeyGenerator func(c *fiber.Ctx) string
	StatusCode   int
	Message      string
	CacheService *CacheService
}

// DefaultRateLimitConfig is the default rate limit config
var DefaultRateLimitConfig = RateLimitConfig{
	Next:         nil,
	Max:          100,
	Duration:     1 * time.Hour,
	KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	StatusCode:   429,
	Message:      "Too many requests",
}

// Helper function to generate session ID
func generateSessionID() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))
}
