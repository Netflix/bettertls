//go:build !go1.18

package test_executor

func GetBuildRevision() string {
	return "build_revision_not_specified"
}
