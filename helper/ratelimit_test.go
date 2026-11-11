package helper

import (
	"testing"
	"time"
)

// TestLoginLimiterBlocksAfterMax memastikan percobaan login IP diblokir
// setelah melewati batas kegagalan, lalu kembali boleh setelah jendela
// waktu lewat.
func TestLoginLimiterBlocksAfterMax(t *testing.T) {
	t.Parallel()
	l := NewLoginLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !l.Allow("10.0.0.1") {
			t.Fatalf("percobaan ke-%d harus masih boleh", i+1)
		}
		l.RecordFailure("10.0.0.1")
	}
	if l.Allow("10.0.0.1") {
		t.Error("setelah 3 kegagalan, IP harus diblokir")
	}
	if !l.Allow("10.0.0.2") {
		t.Error("IP lain tidak boleh ikut terblokir")
	}
}

// TestLoginLimiterResetsOnSuccess memastikan login berhasil me-reset
// penghitung kegagalan untuk IP tersebut.
func TestLoginLimiterResetsOnSuccess(t *testing.T) {
	t.Parallel()
	l := NewLoginLimiter(2, time.Minute)

	l.RecordFailure("10.0.0.1")
	l.RecordFailure("10.0.0.1")
	if l.Allow("10.0.0.1") {
		t.Fatal("IP harus terblokir sebelum reset")
	}
	l.Reset("10.0.0.1")
	if !l.Allow("10.0.0.1") {
		t.Error("setelah reset, IP harus boleh mencoba lagi")
	}
}

// TestLoginLimiterWindowExpires memastikan jendela waktu yang lewat membuka
// kembali akses untuk IP yang sebelumnya diblokir.
func TestLoginLimiterWindowExpires(t *testing.T) {
	t.Parallel()
	l := NewLoginLimiter(1, time.Minute)
	now := time.Now()
	l.now = func() time.Time { return now }

	l.RecordFailure("10.0.0.1")
	if l.Allow("10.0.0.1") {
		t.Fatal("IP harus terblokir dalam jendela aktif")
	}
	now = now.Add(2 * time.Minute)
	if !l.Allow("10.0.0.1") {
		t.Error("setelah jendela lewat, IP harus boleh mencoba lagi")
	}
}

// TestNewLoginLimiterDefaults memastikan nilai tak valid jatuh ke default
// modul (5 kegagalan per menit).
func TestNewLoginLimiterDefaults(t *testing.T) {
	t.Parallel()
	l := NewLoginLimiter(0, 0)
	if l.max != DefaultLoginLimit || l.window != DefaultLoginWindow {
		t.Errorf("max=%d window=%v, mau %d / %v", l.max, l.window, DefaultLoginLimit, DefaultLoginWindow)
	}
}
