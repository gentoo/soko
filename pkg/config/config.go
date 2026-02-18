// SPDX-License-Identifier: GPL-2.0-only
package config

import (
	"fmt"
	"os"
	"time"
)

func PortDir() string {
	return getEnv("SOKO_PORT_DIR", "/mnt/packages-tree/gentoo")
}

func PostgresUser() string {
	return getEnv("SOKO_POSTGRES_USER", "root")
}

func PostgresPass() string {
	return getEnv("SOKO_POSTGRES_PASS", "root")
}

func PostgresDb() string {
	return getEnv("SOKO_POSTGRES_DB", "soko")
}

func PostgresHost() string {
	return getEnv("SOKO_POSTGRES_HOST", "db")
}

func PostgresPort() string {
	return getEnv("SOKO_POSTGRES_PORT", "5432")
}

func Debug() bool {
	return getEnv("SOKO_DEBUG", "false") == "true"
}

func Quiet() bool {
	return getEnv("SOKO_QUIET", "false") == "true"
}

func LogFile() string {
	return getEnv("SOKO_LOG_FILE", "/var/log/soko/errors.log")
}

func Version() string {
	return getEnv("SOKO_VERSION", "v1.0.3")
}

func Port() string {
	return getEnv("SOKO_PORT", "5000")
}

func GithubAPIToken() string {
	return getEnv("SOKO_GITHUB_TOKEN", "")
}

func CodebergAPIToken() string {
	return getEnv("SOKO_CODEBERG_TOKEN", "")
}

func CacheControl() string {
	return getEnv("SOKO_CACHE_CONTROL", "max-age=300")
}

const CacheTime = 5 * time.Minute

func UserAgent() string {
	return fmt.Sprintf("Gentoo Soko %s/packages.gentoo.org/gpackages@gentoo.org", Version())
}

func getEnv(key string, fallback string) string {
	if os.Getenv(key) != "" {
		return os.Getenv(key)
	} else {
		return fallback
	}
}
