package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net"
	"net/http"
	"os"
	"socialNetwork/pkg/db"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type rateBucket struct {
	tokens  float64
	updated time.Time
}

// A bounded token bucket permits normal bursts without unbounded attacker keys.
type rateLimiter struct {
	sync.Mutex
	entries   map[string]rateBucket
	scope     string
	rate      float64
	burst     float64
	nextSweep time.Time
}

func newRateLimiter(scope string, perMinute, burst int) *rateLimiter {
	return &rateLimiter{scope: scope, entries: make(map[string]rateBucket), rate: float64(perMinute) / 60, burst: float64(burst)}
}

func (l *rateLimiter) allow(key string, now time.Time) (bool, int) {
	if db.IsPostgres() {
		return sharedLimit(l.scope, key, l.rate, l.burst)
	}
	l.Lock()
	defer l.Unlock()
	if !now.Before(l.nextSweep) {
		for k, b := range l.entries {
			if now.Sub(b.updated) > 15*time.Minute {
				delete(l.entries, k)
			}
		}
		l.nextSweep = now.Add(time.Minute)
	}
	b, exists := l.entries[key]
	if !exists {
		// Fail closed when full; do not evict an attacker's exhausted bucket.
		if len(l.entries) >= 10000 {
			return false, 60
		}
		b = rateBucket{tokens: l.burst, updated: now}
	}
	if now.Before(b.updated) {
		now = b.updated
	}
	b.tokens = math.Min(l.burst, b.tokens+now.Sub(b.updated).Seconds()*l.rate)
	b.updated = now
	if b.tokens < 1 {
		l.entries[key] = b
		return false, int(math.Ceil((1 - b.tokens) / l.rate))
	}
	b.tokens--
	l.entries[key] = b
	return true, 0
}

var requestLimits = newRateLimiter("requests", 300, 100)
var loginLimits = newRateLimiter("login", 10, 10)
var socketLimits = newRateLimiter("socket", 120, 30)

// Forwarded IPs are accepted only from explicitly configured reverse proxies.
func ClientIP(r *http.Request) string {
	if os.Getenv("VERCEL") != "" {
		raw := strings.TrimSpace(strings.Split(r.Header.Get("X-Vercel-Forwarded-For"), ",")[0])
		if ip := net.ParseIP(raw); ip != nil {
			return ip.String()
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	peer := net.ParseIP(host)
	for _, raw := range strings.Split(os.Getenv("TRUSTED_PROXY_CIDRS"), ",") {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err == nil && network.Contains(peer) {
			if forwarded := net.ParseIP(r.Header.Get("X-Real-IP")); forwarded != nil {
				return forwarded.String()
			}
		}
	}
	return host
}

func AllowedOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	if configured := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/"); configured != "" {
		return origin == configured
	}
	return origin == "http://localhost:3000" || origin == "http://localhost:3001"
}

func AllowSocketMessage(userID int) bool {
	allowed, _ := socketLimits.allow(strconv.Itoa(userID), time.Now())
	return allowed
}

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		if !strings.HasPrefix(r.URL.Path, "/uploads/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		// Enforce origin checks on every route, including WebSocket handshakes.
		if !AllowedOrigin(r.Header.Get("Origin")) {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" && r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "Cross-site request denied", http.StatusForbidden)
			return
		}
		key := ClientIP(r)
		allowed, retry := requestLimits.allow(key, time.Now())
		if allowed && r.Method != http.MethodOptions && (r.URL.Path == "/login" || r.URL.Path == "/register" || (strings.HasPrefix(r.URL.Path, "/auth/google/") && r.URL.Path != "/auth/google/config")) {
			allowed, retry = loginLimits.allow(key, time.Now())
		}
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too many requests. Please wait before trying again."}`))
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		next.ServeHTTP(w, r)
	})
}

var sharedSweep atomic.Int64

func sharedLimit(scope, key string, refill, burst float64) (bool, int) {
	moment := time.Now().Unix()
	last := sharedSweep.Load()
	if moment-last >= 60 && sharedSweep.CompareAndSwap(last, moment) {
		_, _ = db.DBInstance.DB.Exec(`DELETE FROM rate_limits WHERE updated_at<?`, time.Now().Add(-15*time.Minute))
	}
	hash := sha256.Sum256([]byte(key))
	var retry int
	if db.DBInstance.DB.QueryRow(`SELECT take_rate_limit(?,?,?,?)`, scope, hex.EncodeToString(hash[:]), refill, burst).Scan(&retry) != nil {
		return false, 60
	}
	return retry == 0, retry
}
