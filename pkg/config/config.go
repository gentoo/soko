package config

import "os"

func PortDir() string {
	return getEnv("SOKO_PORT_DIR", "/mnt/packages-tree/gentoo")
}

func SelfCheckPortDir() string {
	return getEnv("SOKO_SELFCHECK_PORT_DIR", "/mnt/selfcheck-packages-tree")
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

func DevMode() string {
	return getEnv("SOKO_DEVMODE", "false")
}

func Debug() string {
	return getEnv("SOKO_DEBUG", "false")
}

func Quiet() string {
	return getEnv("SOKO_QUIET", "false")
}

func LogFile() string {
	return getEnv("SOKO_LOG_FILE", "/var/log/soko/errors.log")
}

func Version() string {
	return getEnv("SOKO_VERSION", "v1.0.0")
}

func Port() string {
	return getEnv("SOKO_PORT", "5000")
}

func GithubAPIToken() string {
	return getEnv("SOKO_GITHUB_TOKEN", "")
}

func GraphiqlEndpoint() string {
	return getEnv("GRAPHIQL_ENDPOINT", "https://packages.gentoo.org/api/graphql/")
}

func CacheControl() string {
	return getEnv("SOKO_CACHE_CONTROL", "max-age=300")
}

func getEnv(key string, fallback string) string {
	if os.Getenv(key) != "" {
		return os.Getenv(key)
	} else {
		return fallback
	}
}
