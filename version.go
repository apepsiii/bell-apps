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
	} else {
		os.Setenv("APP_VERSION", Version)
	}
	if v := os.Getenv("BUILD_DATE"); v != "" {
		BuildDate = v
	} else {
		os.Setenv("BUILD_DATE", BuildDate)
	}
	if v := os.Getenv("GIT_COMMIT"); v != "" {
		GitCommit = v
	} else {
		os.Setenv("GIT_COMMIT", GitCommit)
	}
	if v := os.Getenv("APP_NAME"); v != "" {
		AppName = v
	} else {
		os.Setenv("APP_NAME", AppName)
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
