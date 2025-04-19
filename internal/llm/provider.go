package llm

import (
	"github.com/elecmonkey/prompt2cmd/internal/history"
)

// ExecutionAuditResult 命令执行审计结果
type ExecutionAuditResult struct {
	Success     bool   `json:"success"`      // 命令是否成功执行
	Description string `json:"description"`  // 对执行结果的解释
}

// Provider 定义了语言模型提供商的接口
type Provider interface {
	// GenerateCommand 根据提示和上下文生成命令和解释
	// prompt: 用户输入的自然语言
	// historyRecords: 历史记录，包含之前的交互
	// 返回生成的命令、命令解释和可能的错误
	GenerateCommand(prompt string, historyRecords []history.HistoryRecord) (command string, explanation string, err error)
	
	// AuditExecutionResult 审计命令执行结果
	// command: 执行的命令
	// result: 命令执行结果
	// prompt: 用户的原始需求
	// 返回审计结果和可能的错误
	AuditExecutionResult(command string, result string, prompt string) (*ExecutionAuditResult, error)
	
	// IsLocal 返回是否为本地模型
	IsLocal() bool
} 