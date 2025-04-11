package helpers

import "regexp"

// Benchmark 压测结果
// goos: darwin
// goarch: amd64
// pkg: code.byted.org/security/go-polaris/helpers
//
// BenchmarkAllow
// BenchmarkAllow          	250605328	         4.57 ns/op
// BenchmarkAllow-2        	252492027	         5.31 ns/op
// BenchmarkAllow-4        	253907700	         4.91 ns/op
// BenchmarkAllow-8        	256431940	         4.76 ns/op
// BenchmarkAllow-12       	251342673	         4.82 ns/op
//
// BenchmarkAllowReg
// BenchmarkAllowReg       	  388852	      2731 ns/op
// BenchmarkAllowReg-2     	  439770	      2589 ns/op
// BenchmarkAllowReg-4     	  464236	      2633 ns/op
// BenchmarkAllowReg-8     	  463546	      2654 ns/op
// BenchmarkAllowReg-12    	  455108	      2602 ns/op

//Allow 判断数据是否为名单内数据
func Allow(data string, allow ...string) bool {
	return contains(data, allow...)
}

//AllowReg 判断数据是否符合正则表达式
func AllowReg(data string, patterns ...string) bool {
	for _, pattern := range patterns {
		if ok, _ := regexp.MatchString(pattern, data); ok {
			return true
		}
	}

	return false
}

//NewAllowList 新建allow list结构体
func NewAllowList(allow ...string) *AllowList {
	a := new(AllowList)
	a.SetList(allow...)
	return a
}

//AllowList 过滤列表
type AllowList struct {
	list []string
	regs []*regexp.Regexp
}

//Allow 判断数据是否在allow list中
func (a *AllowList) Allow(data string) bool {
	if contains(data, a.list...) {
		return true
	}

	for _, reg := range a.regs {
		if reg.MatchString(data) {
			return true
		}
	}

	return false
}

//SetList 配置允许条目清单
func (a *AllowList) SetList(allow ...string) {
	a.list = make([]string, len(allow))
	copy(a.list, allow)
}

//GetList 获取列表
func (a *AllowList) GetList() []string {
	list := make([]string, len(a.list))
	copy(list, a.list)
	return list
}

//AppendList 追加列表
func (a *AllowList) AppendList(allow ...string) {
	a.list = append(a.list, allow...)
}

//SetRegs 配置允许正则列表
func (a *AllowList) SetRegs(patterns ...string) {
	a.regs = make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		a.regs = append(a.regs, reg)
	}
}

//GetRegs 获取正则列表
func (a *AllowList) GetRegs() []*regexp.Regexp {
	regs := make([]*regexp.Regexp, len(a.regs))
	copy(regs, a.regs)
	return regs
}

//AppendReg 追加正则
func (a *AllowList) AppendReg(patterns ...string) {
	for _, pattern := range patterns {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		a.regs = append(a.regs, reg)
	}
}

//Reset 清空已有列表
func (a *AllowList) Reset() {
	a.list = nil
	a.regs = nil
}

func contains(data string, allow ...string) bool {
	for _, item := range allow {
		if item == data {
			return true
		}
	}
	return false
}
