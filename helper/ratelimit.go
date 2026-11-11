package helper

import (
	"sync"
	"time"
)

// Batas default rate limit login (Modul 12): "Dibatasi 5 kali gagal per
// menit per IP". Penyimpanan in-memory: Redis di repo ini hanya dipakai
// worker/relay (AGENTS.md §4), jadi limiter ini sengaja ringan dan
// proses-lokal.
const (
	DefaultLoginLimit  = 5
	DefaultLoginWindow = time.Minute
)

// loginAttempt menyimpan jumlah kegagalan login pada satu IP dalam satu
// jendela waktu.
type loginAttempt struct {
	failures    int
	windowStart time.Time
}

// LoginLimiter membatasi percobaan login gagal per IP.
//
// Sifat penting:
//   - hanya kegagalan kredensial yang dihitung (password/username salah);
//   - login berhasil me-reset penghitung untuk IP tersebut;
//   - jendela waktu dimulai dari kegagalan pertama; setelah jendela lewat,
//     IP dianggap bersih lagi.
type LoginLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	attempts map[string]loginAttempt
	now      func() time.Time
}

// NewLoginLimiter membuat limiter dengan batas maksimum kegagalan dan
// panjang jendela waktu. Nilai nol/negatif memakai nilai default.
func NewLoginLimiter(max int, window time.Duration) *LoginLimiter {
	if max <= 0 {
		max = DefaultLoginLimit
	}
	if window <= 0 {
		window = DefaultLoginWindow
	}
	return &LoginLimiter{
		max:      max,
		window:   window,
		attempts: map[string]loginAttempt{},
		now:      time.Now,
	}
}

// DefaultLoginLimiter mengembalikan limiter bersama yang dipakai AuthService
// bila tidak ada limiter yang diinjeksi (mis. pada test).
func DefaultLoginLimiter() *LoginLimiter {
	defaultLimiterOnce.Do(func() {
		defaultLoginLimiter = NewLoginLimiter(DefaultLoginLimit, DefaultLoginWindow)
	})
	return defaultLoginLimiter
}

var (
	defaultLimiterOnce  sync.Once
	defaultLoginLimiter *LoginLimiter
)

// Allow memberi tahu apakah IP masih boleh mencoba login.
func (l *LoginLimiter) Allow(ip string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.attempts[ip]
	if !ok {
		return true
	}
	if l.now().Sub(a.windowStart) >= l.window {
		return true
	}
	return a.failures < l.max
}

// RecordFailure mencatat satu kegagalan login untuk IP.
func (l *LoginLimiter) RecordFailure(ip string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune()
	now := l.now()
	a, ok := l.attempts[ip]
	if !ok || now.Sub(a.windowStart) >= l.window {
		l.attempts[ip] = loginAttempt{failures: 1, windowStart: now}
		return
	}
	a.failures++
	l.attempts[ip] = a
}

// Reset menghapus penghitung kegagalan untuk IP (dipanggil setelah login
// berhasil).
func (l *LoginLimiter) Reset(ip string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// prune membuang entri yang jendelanya sudah lewat. Dipanggil dengan lock
// dipegang supaya map tidak tumbuh tanpa batas.
func (l *LoginLimiter) prune() {
	now := l.now()
	for ip, a := range l.attempts {
		if now.Sub(a.windowStart) >= l.window {
			delete(l.attempts, ip)
		}
	}
}
