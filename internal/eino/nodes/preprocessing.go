// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"
	"strings"
	"unicode"
)

// PreprocessInput 预处理输入
type PreprocessInput struct {
	Query    string
	UserType string
}

// PreprocessOutput 预处理输出
type PreprocessOutput struct {
	Query    string
	UserType string
}

// PreprocessQuery 查询预处理 Lambda 函数
// 对用户查询进行标准化处理
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

// PreprocessQueryToString 预处理查询并返回字符串
// 用于简单的 Graph 流程
func PreprocessQueryToString(ctx context.Context, query string) (string, error) {
	// 1. 去除首尾空白
	query = strings.TrimSpace(query)

	// 2. 规范化空白字符
	query = normalizeWhitespace(query)

	// 3. 移除特殊控制字符
	query = removeControlChars(query)

	return query, nil
}

// normalizeWhitespace 规范化空白字符
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// removeControlChars 移除控制字符
func removeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
}

// removeSpecialChars 移除特殊字符（可选使用）
func removeSpecialChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) || r == '?' || r == '？' {
			return r
		}
		return -1
	}, s)
}

// NormalizeQuery 标准化查询（更激进的处理）
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
