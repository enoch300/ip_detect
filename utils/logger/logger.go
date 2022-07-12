/*
* @Author: wangqilong
* @Description:
* @File: log
* @Date: 2021/9/15 11:05 下午
 */

package logger

import (
	"github.com/enoch300/glog"
	"github.com/sirupsen/logrus"
	"ip_detect/utils"
	"path/filepath"
	"time"
)

var Global *logrus.Logger

func InitLogger() {
	logPath := filepath.Dir(utils.GetCurrentAbPath()) + "/logs"
	Global = glog.NewLogger(logPath, "ipdetect", "_%Y-%m-%d.log", 72*time.Hour, 24*time.Hour)
}
