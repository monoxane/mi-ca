package utils

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func GetImageName(container corev1.Container) string {
	fromDockerHub := true
	namespace := ""

	splitBySlash := strings.Split(container.Image, "/")

	if len(splitBySlash) == 3 {
		fromDockerHub = false
	}

	if fromDockerHub {
		namespace = splitBySlash[0]
	} else {
		namespace = splitBySlash[2]
	}

	namespace = strings.Split(namespace, ":")[0]

	return namespace
}
