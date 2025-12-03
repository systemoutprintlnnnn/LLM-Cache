package status

// StatusCode 定义统一的业务状态码类型。
// 说明：尽量保持简单以满足当前项目使用场景。
// 0 表示成功，其余为错误状态。
type StatusCode int

const (
	// CodeOK 操作成功。
	CodeOK StatusCode = 0

	// ErrCodeInvalidParam 参数验证失败或格式错误。
	ErrCodeInvalidParam StatusCode = 1001
	// ErrCodeInternal 服务器内部错误。
	ErrCodeInternal StatusCode = 1002
	// ErrCodeUnavailable 服务暂时不可用。
	ErrCodeUnavailable StatusCode = 1003
	// ErrCodeNotFound 请求的资源不存在。
	ErrCodeNotFound StatusCode = 1004
)

// String 返回状态码对应的字符串描述。
// 用于日志记录和错误信息展示。
func (c StatusCode) String() string {
	switch c {
	case CodeOK:
		return "OK"
	case ErrCodeInvalidParam:
		return "INVALID_PARAM"
	case ErrCodeInternal:
		return "INTERNAL_ERROR"
	case ErrCodeUnavailable:
		return "UNAVAILABLE"
	case ErrCodeNotFound:
		return "NOT_FOUND"
	default:
		return "UNKNOWN"
	}
}
