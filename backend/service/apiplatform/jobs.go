package apiplatform

import (
	"fmt"
	"strings"
	"time"

	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/utils/logger"
)

// PullChannels 按筛选条件预览上游渠道，不入库。
func PullChannels(base, token, keyword, status string, minID, maxID int) ([]RemoteChannel, float64, error) {
	perUnit, err := QuotaPerUnit(base, token)
	if err != nil {
		return nil, 0, err
	}
	rows, err := SearchChannels(base, token, keyword, status)
	if err != nil {
		return nil, 0, err
	}
	return FilterChannels(rows, keyword, minID, maxID, status), perUnit, nil
}

// SyncPlatform 用渠道搜索一次拉回用量，再更新已保存的渠道。
func SyncPlatform(platform *model.ApiPlatform) (updated, failed int, err error) {
	channels, err := model.ListApiPlatformChannels(platform.Id)
	if err != nil {
		return 0, 0, err
	}
	if len(channels) == 0 {
		return 0, 0, nil
	}
	perUnit, err := QuotaPerUnit(platform.BaseURL, platform.APIKey)
	if err != nil {
		return 0, 0, err
	}
	remotes, err := SearchChannels(platform.BaseURL, platform.APIKey, "", "")
	if err != nil {
		return 0, 0, err
	}
	byID := make(map[int]RemoteChannel, len(remotes))
	for _, row := range remotes {
		byID[row.ID] = row
	}
	updates := make([]model.ChannelUsageUpdate, 0, len(channels))
	for _, ch := range channels {
		remote, ok := byID[ch.ChannelId]
		if !ok {
			failed++
			logger.SysLog(fmt.Sprintf("api platform %d channel %d sync skipped: not in channel list", platform.Id, ch.ChannelId))
			continue
		}
		updates = append(updates, model.ChannelUsageUpdate{
			RowID:         ch.Id,
			Name:          remote.Name,
			Status:        remote.Status,
			UsedQuota:     remote.UsedQuota,
			UsedAmount:    QuotaToUSD(remote.UsedQuota, perUnit),
			SettledAmount: ch.SettledAmount,
		})
	}
	if err := model.ApplyApiPlatformChannelUsages(platform.Id, updates); err != nil {
		return 0, failed, err
	}
	return len(updates), failed, nil
}

// SyncAll 同步全部台子。
func SyncAll() {
	rows, err := model.ListApiPlatforms()
	if err != nil {
		logger.SysLog("api platform sync list failed: " + err.Error())
		return
	}
	for i := range rows {
		updated, failed, err := SyncPlatform(&rows[i])
		if err != nil {
			logger.SysLog(fmt.Sprintf("api platform %d sync failed: %s", rows[i].Id, err.Error()))
			continue
		}
		logger.SysLog(fmt.Sprintf("api platform %d sync updated=%d failed=%d", rows[i].Id, updated, failed))
	}
}

// DiscoverPlatform 按已保存的筛选名称发现新渠道，并测试模型。
func DiscoverPlatform(platform *model.ApiPlatform) error {
	names, err := model.DistinctFilterNames(platform.Id)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	settings, err := model.GetAPIPlatformSettings()
	if err != nil {
		return err
	}
	perUnit, err := QuotaPerUnit(platform.BaseURL, platform.APIKey)
	if err != nil {
		return err
	}
	existing, err := model.ExistingUpstreamIDs(platform.Id)
	if err != nil {
		return err
	}
	var created []model.ApiPlatformChannel
	for _, name := range names {
		rows, err := SearchChannels(platform.BaseURL, platform.APIKey, name, "")
		if err != nil {
			logger.SysLog(fmt.Sprintf("api platform %d discover %s failed: %s", platform.Id, name, err.Error()))
			continue
		}
		rows = FilterChannels(rows, name, 0, 0, "")
		for _, row := range rows {
			if _, ok := existing[row.ID]; ok {
				continue
			}
			now := time.Now()
			used := QuotaToUSD(row.UsedQuota, perUnit)
			incoming := model.ApiPlatformChannel{
				PlatformId:      platform.Id,
				ChannelId:       row.ID,
				Name:            row.Name,
				FilterName:      name,
				Models:          row.Models,
				Status:          row.Status,
				UsedQuota:       row.UsedQuota,
				UsedAmount:      used,
				SettledAmount:   0,
				UnsettledAmount: used,
				LastSyncAt:      &now,
			}
			if err := model.SavePulledChannel(platform.Id, incoming); err != nil {
				logger.SysLog(fmt.Sprintf("api platform %d save new channel %d failed: %s", platform.Id, row.ID, err.Error()))
				continue
			}
			saved, err := findChannel(platform.Id, row.ID)
			if err != nil {
				logger.SysLog(fmt.Sprintf("api platform %d reload channel %d failed: %s", platform.Id, row.ID, err.Error()))
				continue
			}
			existing[row.ID] = *saved
			created = append(created, *saved)
		}
	}
	for _, ch := range created {
		summary := testAndStripModels(platform, &ch, settings.TestModels)
		_ = model.UpdateApiPlatformChannelTest(platform.Id, ch.Id, summary)
	}
	return model.RecalcApiPlatform(platform.Id)
}

// DiscoverAll 发现全部台子的新渠道。
func DiscoverAll() {
	rows, err := model.ListApiPlatforms()
	if err != nil {
		logger.SysLog("api platform discover list failed: " + err.Error())
		return
	}
	for i := range rows {
		if err := DiscoverPlatform(&rows[i]); err != nil {
			logger.SysLog(fmt.Sprintf("api platform %d discover failed: %s", rows[i].Id, err.Error()))
		}
	}
}

// TestSavedChannel 手动测试一个模型，只返回结果。
func TestSavedChannel(platform *model.ApiPlatform, channel *model.ApiPlatformChannel, modelName string) (int, string, error) {
	code, _, message, err := TestChannel(platform.BaseURL, platform.APIKey, channel.ChannelId, modelName)
	if err != nil {
		return 0, "", err
	}
	text := fmt.Sprintf("%s: %d %s", modelName, code, message)
	_ = model.UpdateApiPlatformChannelTest(platform.Id, channel.Id, text)
	return code, message, nil
}

// ReplaceChannelModels 上游更新成功后再写本地模型。
func ReplaceChannelModels(platform *model.ApiPlatform, channel *model.ApiPlatformChannel, models []string) (string, error) {
	cleaned := make([]string, 0, len(models))
	seen := map[string]struct{}{}
	for _, name := range models {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		cleaned = append(cleaned, name)
	}
	joined := strings.Join(cleaned, ",")
	if err := UpdateChannelModels(platform.BaseURL, platform.APIKey, channel.ChannelId, joined); err != nil {
		return "", err
	}
	if err := model.UpdateApiPlatformChannelModels(platform.Id, channel.Id, joined); err != nil {
		return "", err
	}
	return joined, nil
}

// SetSavedChannelStatus 先改上游状态，成功后再写本地。
func SetSavedChannelStatus(platform *model.ApiPlatform, channel *model.ApiPlatformChannel, status int) error {
	if status != 1 && status != 2 {
		return fmt.Errorf("不支持的状态")
	}
	if err := UpdateChannelStatus(platform.BaseURL, platform.APIKey, channel.ChannelId, status); err != nil {
		return err
	}
	return model.UpdateApiPlatformChannelStatus(platform.Id, channel.Id, status)
}

func testAndStripModels(platform *model.ApiPlatform, channel *model.ApiPlatformChannel, models []string) string {
	current := channel.Models
	remote, err := GetChannel(platform.BaseURL, platform.APIKey, channel.ChannelId)
	if err == nil {
		current = remote.Models
	}
	lines := make([]string, 0, len(models))
	removedAny := false
	for _, name := range models {
		code, drop, message, err := TestChannel(platform.BaseURL, platform.APIKey, channel.ChannelId, name)
		if err != nil {
			lines = append(lines, fmt.Sprintf("%s: 请求失败 %s", name, err.Error()))
			continue
		}
		if !drop {
			lines = append(lines, fmt.Sprintf("%s: %d %s", name, code, message))
			continue
		}
		next, removed := RemoveModel(current, name)
		if !removed {
			lines = append(lines, fmt.Sprintf("%s: %d 不在模型列表", name, code))
			continue
		}
		current = next
		removedAny = true
		lines = append(lines, fmt.Sprintf("%s: %d 已移除", name, code))
	}
	if removedAny {
		if err := UpdateChannelModels(platform.BaseURL, platform.APIKey, channel.ChannelId, current); err != nil {
			lines = append(lines, "更新上游模型失败: "+err.Error())
		} else if err := model.UpdateApiPlatformChannelModels(platform.Id, channel.Id, current); err != nil {
			lines = append(lines, "更新本地模型失败: "+err.Error())
		}
	}
	return strings.Join(lines, "\n")
}

func findChannel(platformID, upstreamID int) (*model.ApiPlatformChannel, error) {
	rows, err := model.ListApiPlatformChannels(platformID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].ChannelId == upstreamID {
			return &rows[i], nil
		}
	}
	return nil, fmt.Errorf("渠道 %d 未找到", upstreamID)
}
