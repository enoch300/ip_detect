/*
* @Author: wangqilong
* @Description:
* @File: strToFloat64
* @Date: 2021/7/7 8:18 下午
 */

package conv

import (
	"fmt"
	"strconv"
)

func FormatFloat64(f float64) float64 {
	v, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", f), 64)
	return v
}
