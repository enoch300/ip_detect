<<<<<<< HEAD
/*
* @Author: wangqilong
* @Description:
* @File: log
* @Date: 2021/9/15 11:05 下午
 */

=======
>>>>>>> 52c810ff9b823c0b781aeb745e80364c1b3b7a8b
package log

import (
	"github.com/enoch300/glog"
	"github.com/sirupsen/logrus"
<<<<<<< HEAD
	"ip_detect/utils"
	"path/filepath"
=======
	"path/filepath"
	"xh_detect/utils"
>>>>>>> 52c810ff9b823c0b781aeb745e80364c1b3b7a8b
)

var GlobalLog *logrus.Logger

func NewLogger(save uint) {
	logPath := filepath.Dir(utils.GetCurrentAbPath()) + "/logs"
<<<<<<< HEAD
	GlobalLog = glog.NewLogger(logPath, "ip_detect", save)
=======
	GlobalLog = glog.NewLogger(logPath, "xh_detect", save)
>>>>>>> 52c810ff9b823c0b781aeb745e80364c1b3b7a8b
}
