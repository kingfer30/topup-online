package constants

import (
	"github.com/kingfer30/topup-online/utils/env"
)

// GetServerPort 延迟读取端口号，确保能读取到 .env 中的配置
func GetServerPort() string {
	return env.String("PORT", "3030")
}

// 其他环境变量也建议改为函数调用以保证一致性
func GetDebugEnabled() bool {
	return env.Bool("DEBUG", false)
}

func GetDebugSQLEnabled() bool {
	return env.Bool("DEBUG_SQL", false)
}

func GetRelayProxy() string {
	return env.String("RELAY_PROXY", "")
}

func GetRelayTimeout() int {
	return env.Int("RELAY_TIMEOUT", 0)
}

func GetCacheFrequency() int {
	return env.Int("CACHE_FREQUENCY", 300)
}

// GetDataDir 获取数据目录路径，用于存放 .initialized、.env 等持久化文件
// Docker 部署时设置 DATA_DIR=/app/data 并挂载该目录即可持久化
func GetDataDir() string {
	return env.String("DATA_DIR", ".")
}

// common
const (
	LoggerDEBUG    = "DEBUG"
	LoggerINFO     = "INFO"
	LoggerWarn     = "WARN"
	LoggerError    = "ERR"
	RequestIdKey   = "X-Guoguo-Request-Id"
	KeyRequestBody = "key_request_body"
)

// db
const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100

	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
	UserStatusDeleted  = 3
)
