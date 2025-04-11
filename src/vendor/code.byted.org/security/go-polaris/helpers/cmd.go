package helpers

import (
	"regexp"
)

var (
	reg *regexp.Regexp
)

func init() {
	reg, _ = regexp.Compile(`^[a-z0-9A-Z\-_:/?=@.,]*$`)
}

// CheckCmdParam check if param invalid 检查命令行参数是否合法
// return true if param is safe, otherwise return false
// 如果参数安全返回true，如果参数不安全返回false
func CheckCmdParam(param string) bool {
	return reg.MatchString(param)
}

// CheckCmdParams check param list
func CheckCmdParams(params ...string) bool {
	for _, param := range params {
		if !CheckCmdParam(param) {
			return false
		}
	}
	return true
}
