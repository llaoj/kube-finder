package kube

import (
	"regexp"
)

func ParseContainerID(containerID string) string {
	pattern := regexp.MustCompile(`[0-9a-f]{64}`)
	parts := pattern.FindStringSubmatch(containerID)
	if parts != nil {
		return parts[0]
	}

	return ""
}
