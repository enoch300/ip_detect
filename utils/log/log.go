package log

import (
	"github.com/enoch300/glog"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"xh_detect/utils"
)

var GlobalLog *logrus.Logger

func NewLogger(save uint) {
	logPath := filepath.Dir(utils.GetCurrentAbPath()) + "/logs"
	GlobalLog = glog.NewLogger(logPath, "xh_detect", save)
}
