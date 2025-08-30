package constants

import (
	"os"

	"github.com/kingfer30/topup-online/utils/env"
)

var ServerPort = env.String(os.Getenv("PORT"), "3030")
var DebugEnabled = env.Bool(os.Getenv("DEBUG"), false)
var DebugSQLEnabled = env.Bool(os.Getenv("DEBUG_SQL"), false)
var RelayProxy = env.String(os.Getenv("RELAY_PROXY"), "")
var RelayTimeout = env.Int("RELAY_TIMEOUT", 0)     // unit is second
var CacheFrequency = env.Int("CACHE_FREQUENCY", 300) // unit is second

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
