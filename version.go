package main

// buildVersion is set at build time via -ldflags "-X main.buildVersion=...".
var buildVersion = "dev"

// buildCommit is set at build time via -ldflags "-X main.buildCommit=...".
var buildCommit = "unknown"

// buildDate is set at build time via -ldflags "-X main.buildDate=...".
var buildDate = "unknown"

func versionInfo() map[string]string {
	return map[string]string{
		"version": buildVersion,
		"commit":  buildCommit,
		"date":    buildDate,
	}
}
