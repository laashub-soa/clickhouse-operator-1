package version

import (
	"github.com/sirupsen/logrus"
	"runtime"
)

var (
	Version = "UNKNOWN"
)

func PrintVersion() {
	logrus.Infof("Operator Version: %s", Version)
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
}
