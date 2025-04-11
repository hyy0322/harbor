package report

// Reporter 模块使用上报服务
type Reporter struct {
	Module string
	state  int32
}

// NewReporter 返回一个 Reporter
func NewReporter(name string) *Reporter {
	return &Reporter{Module: name}
}

// Deprecated: Report 上传非敏感使用数据，用于后续安全维护。
func (r *Reporter) Report() {
	return
}
