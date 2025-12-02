// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"
	"strings"
	"unicode"
)

// PreprocessInput 定义查询预处理的输入参数。
// 包含原始查询字符串和用户类型标识。
type PreprocessInput struct {
	Query    string
	UserType string
}

// PreprocessOutput 定义查询预处理的输出结果。
// 包含处理后的查询字符串和用户类型标识。
type PreprocessOutput struct {
	Query    string
	UserType string
}

// PreprocessQuery 对查询请求进行预处理 Lambda 函数。
// 执行去除首尾空格、规范化空白字符以及移除控制字符等操作，以提高后续检索的准确性。
// 参数 ctx: 上下文对象。
// 参数 input: 预处理输入对象。
// 返回: 预处理输出对象或错误。
func PreprocessQuery(ctx context.Context, input *PreprocessInput) (*PreprocessOutput, error) {
	query := input.Query

	// 1. 去除首尾空白
	query = strings.TrimSpace(query)

	// 2. 规范化空白字符（多个空格合并为一个）
	query = normalizeWhitespace(query)

	// 3. 移除特殊控制字符
	query = removeControlChars(query)

	return &PreprocessOutput{
		Query:    query,
		UserType: input.UserType,
	}, nil
}

// PreprocessQueryToString 预处理查询并直接返回字符串结果。
// 这是一个简化的版本，适用于不需要传递额外上下文（如 UserType）的 Graph 流程。
// 参数 ctx: 上下文对象。
// 参数 query: 原始查询字符串。
// 返回: 处理后的字符串或错误。
func PreprocessQueryToString(ctx context.Context, query string) (string, error) {
	// 1. 去除首尾空白
	query = strings.TrimSpace(query)

	// 2. 规范化空白字符
	query = normalizeWhitespace(query)

	// 3. 移除特殊控制字符
	query = removeControlChars(query)

	return query, nil
}

// normalizeWhitespace 规范化空白字符，将连续的空白字符替换为单个空格。
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// removeControlChars 移除字符串中的不可打印控制字符（保留换行和制表符）。
func removeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
}

// removeSpecialChars 移除特殊字符，仅保留字母、数字、空格和问号。
// 该函数目前作为可选的辅助工具，用于更激进的文本清洗。
func removeSpecialChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) || r == '?' || r == '？' {
			return r
		}
		return -1
	}, s)
}

// NormalizeQuery 执行更激进的查询标准化处理。
// 除了基础清洗外，还包括转小写和移除大部分标点符号。
// 参数 ctx: 上下文对象。
// 参数 query: 原始查询字符串。
// 返回: 标准化后的字符串或错误。
func NormalizeQuery(ctx context.Context, query string) (string, error) {
	// 1. 基础处理
	query = strings.TrimSpace(query)
	query = normalizeWhitespace(query)

	// 2. 转小写（对于英文查询）
	query = strings.ToLower(query)

	// 3. 移除标点符号（保留问号）
	query = removeSpecialChars(query)

	return query, nil
}
