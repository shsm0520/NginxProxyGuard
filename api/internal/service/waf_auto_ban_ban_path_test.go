package service

import (
	"os"
	"strings"
	"testing"
)

func TestWAFAutoBanUsesRateLimitRepositoryBanPath(t *testing.T) {
	srcBytes, err := os.ReadFile("waf_auto_ban.go")
	if err != nil {
		t.Fatalf("read waf_auto_ban.go: %v", err)
	}
	src := string(srcBytes)
	start := strings.Index(src, "func (s *WAFAutoBanService) banIP(")
	if start < 0 {
		t.Fatal("banIP function not found")
	}
	end := strings.Index(src[start:], "\n// updateNginxDenyList")
	if end < 0 {
		t.Fatal("banIP function end marker not found")
	}
	body := src[start : start+end]

	if strings.Contains(body, "INSERT INTO banned_ips") || strings.Contains(body, "ExecContext(ctx, query") {
		t.Fatalf("WAFAutoBanService.banIP still writes banned_ips directly; it must use RateLimitRepository.BanIP so active-ban cache invalidation stays in one path")
	}
	if !strings.Contains(body, "s.rateLimitRepo.BanAutoIP(ctx, nil, ip, reason, durationSeconds, failCount)") {
		t.Fatalf("WAFAutoBanService.banIP must call s.rateLimitRepo.BanAutoIP(ctx, nil, ip, reason, durationSeconds, failCount); body was:\n%s", body)
	}
}
