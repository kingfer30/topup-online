package scheduler

import (
	"sync"
	"time"

	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/service/apiplatform"
	"github.com/kingfer30/topup-online/utils/logger"
)

// StartAPIPlatformSchedulers 每分钟检查一次基础设置里的开关和间隔。
func StartAPIPlatformSchedulers() {
	var mu sync.Mutex
	var lastSync time.Time
	var lastDiscover time.Time
	running := false

	run := func() {
		mu.Lock()
		if running {
			mu.Unlock()
			return
		}
		running = true
		mu.Unlock()
		defer func() {
			mu.Lock()
			running = false
			mu.Unlock()
		}()

		cfg, err := model.GetAPIPlatformSettings()
		if err != nil {
			logger.SysLog("api platform settings: " + err.Error())
			return
		}
		now := time.Now()
		if cfg.SyncEnabled && due(lastSync, cfg.SyncMinutes, now) {
			lastSync = now
			logger.SysLog("api platform sync started")
			apiplatform.SyncAll()
		}
		if cfg.DiscoverEnabled && due(lastDiscover, cfg.DiscoverMinutes, now) {
			lastDiscover = now
			logger.SysLog("api platform discover started")
			apiplatform.DiscoverAll()
		}
	}

	go run()
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			run()
		}
	}()
}

func due(last time.Time, minutes int, now time.Time) bool {
	if minutes < 1 {
		minutes = 30
	}
	return last.IsZero() || now.Sub(last) >= time.Duration(minutes)*time.Minute
}
