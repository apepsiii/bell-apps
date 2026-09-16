package main

import "os"

var (
	Version   = "2.1.0"
	BuildDate = "unknown"
	GitCommit = "unknown"
	AppName   = "NIBA SuperApps"
)

func init() {
	if v := os.Getenv("APP_VERSION"); v != "" {
		Version = v
	}
	if v := os.Getenv("BUILD_DATE"); v != "" {
		BuildDate = v
	}
	if v := os.Getenv("GIT_COMMIT"); v != "" {
		GitCommit = v
	}
	if v := os.Getenv("APP_NAME"); v != "" {
		AppName = v
	}
}

func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"build_date": BuildDate,
		"git_commit": GitCommit,
		"app_name":   AppName,
	}
}

func GetVersionString() string {
	return "v" + Version
}

func GetFullVersionString() string {
	return AppName + " v" + Version + " (build " + BuildDate + ")"
}
