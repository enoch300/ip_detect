/*
* @Author: wangqilong
* @Description:
* @File: log
* @Date: 2021/9/15 11:05 下午
 */

package log

import (
	"github.com/enoch300/glog"
	"github.com/sirupsen/logrus"
	"ip_detect/utils"
	"path/filepath"
)

var GlobalLog *logrus.Logger

func NewLogger(save uint) {
	logPath := filepath.Dir(utils.GetCurrentAbPath()) + "/logs"
	GlobalLog = glog.NewLogger(logPath, "ip_detect", save)
}
