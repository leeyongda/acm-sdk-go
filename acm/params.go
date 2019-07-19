package acm

import (
	"regexp"
	"strings"
)

var (
	// VALIDCHAR ...
	VALIDCHAR = map[string]string{
		"_": "_",
		"-": "-",
		".": ".",
		":": ":",
	}
	// IsAbc ... 判断是否为字母
	IsAbc = regexp.MustCompile("[a-zA-Z]+")
	// IsNum ... 判断是否为数字
	IsNum = regexp.MustCompile("^[0-9]*$")
)

// IsValid ....
func IsValid(param string) bool {
	if param == "" {
		return false
	}
	for _, i := range param {
		_, ok := VALIDCHAR[string(i)]
		// 条件判断错误，导致panic。
		if IsAbc.MatchString(string(i)) || IsNum.MatchString(string(i)) || ok == true {
			continue
		}
		return false
	}
	return true
}

// CheckParams ...
func CheckParams(p Params) bool {
	ok := IsValid(p.Group)
	isok := IsValid(p.NameSpace)
	if ok && !isok {
		return false
	}
	return true
}

// GroupKey ...
func GroupKey(dataID, group, timestamp string) string {
	data := []string{dataID, group, timestamp}
	return strings.Join(data, "+")
}

// ParseKey ...
func ParseKey(key string) (string, string, string) {
	sp := strings.Split(key, "+")
	return sp[0], sp[1], sp[2]

}
