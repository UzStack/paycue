package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/repository"
	"go.uber.org/zap"
)

// StartStatsReporter bu instance'ning anonim foydalanish statistikasini davriy ravishda
// kollektorga (cfg.StatsURL) yuborib turadi. Maxfiy ma'lumot yuborilmaydi — faqat sanoqlar.
func StartStatsReporter(ctx context.Context, db *sql.DB, log *zap.Logger, cfg *config.Config, version string) {
	instanceID, err := repository.GetOrCreateInstanceID(db)
	if err != nil {
		log.Info("stats: instance id olinmadi", zap.Error(err))
		return
	}
	go func() {
		// Server tiklansin deb biroz kutamiz, keyin har 6 soatda yuboramiz.
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				sendStatsReport(ctx, db, log, cfg, version, instanceID)
				timer.Reset(6 * time.Hour)
			}
		}
	}()
}

func sendStatsReport(ctx context.Context, db *sql.DB, log *zap.Logger, cfg *config.Config, version, instanceID string) {
	r := repository.LocalStats(db)
	r.InstanceID = instanceID
	r.Version = version
	r.OS = runtime.GOOS
	r.Arch = runtime.GOARCH

	payload, err := json.Marshal(r)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.StatsURL+"/api/stats/report", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Info("stats: hisobot yuborilmadi", zap.String("url", cfg.StatsURL), zap.Error(err))
		return
	}
	resp.Body.Close()
	log.Info("stats: hisobot yuborildi", zap.String("url", cfg.StatsURL), zap.Int("status", resp.StatusCode))
}
