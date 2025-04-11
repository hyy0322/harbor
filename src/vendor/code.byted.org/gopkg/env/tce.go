package env

import (
	"os"
)

var (
	podName        string
	imageVersion   string
	cloudEnv       string
	inVolcanoCloud bool
)

func init() {
	podName = os.Getenv("MY_POD_NAME")
	if podName == "" {
		podName = "-"
	}
	imageVersion = os.Getenv("CURRENT_VERSION")
	if imageVersion == "" {
		imageVersion = "-"
	}
	cloudEnv = os.Getenv("PAAS_CLOUD_ENV")
	if cloudEnv == "" {
		cloudEnv = "-"
	}
	if cloudEnv == "VOLCANO" {
		inVolcanoCloud = true
	}
}

func InTCE() bool {
	return inTCE
}

func PodName() string {
	return podName
}

func ImageVersion() string {
	return imageVersion
}

func CloudEnv() string {
	return cloudEnv
}

func InVolcanoCloud() bool {
	return inVolcanoCloud
}
