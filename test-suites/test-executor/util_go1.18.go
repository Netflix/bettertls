//go:build go1.18

package test_executor

import (
	"runtime/debug"
)

func GetBuildRevision() string {
	if debugInfo, ok := debug.ReadBuildInfo(); ok {
		var revision string
		var dirty bool
		for _, setting := range debugInfo.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
			}
			if setting.Key == "vcs.modified" {
				dirty = setting.Value == "true"
			}
		}
		if revision != "" && dirty {
			revision += "-dirty"
		}
		return revision
	}
	return ""
}
