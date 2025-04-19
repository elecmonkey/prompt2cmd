package security

import (
	"strings"
)

// SecurityChecker 安全检查器接口
type SecurityChecker interface {
	// IsDangerousCommand 检查命令是否危险
	IsDangerousCommand(command string) bool
	
	// GetWarningMessage 获取警告信息
	GetWarningMessage(command string) string
}

// DefaultSecurityChecker 默认安全检查器实现
type DefaultSecurityChecker struct {
	DangerousCommands []string
}

// NewSecurityChecker 创建一个新的安全检查器
func NewSecurityChecker(dangerousCommands []string) *DefaultSecurityChecker {
	return &DefaultSecurityChecker{
		DangerousCommands: dangerousCommands,
	}
}

// IsDangerousCommand 检查命令是否危险
func (c *DefaultSecurityChecker) IsDangerousCommand(command string) bool {
	if command == "" {
		return false
	}

	// 转换为小写以进行不区分大小写的比较
	cmdLower := strings.ToLower(command)
	
	for _, dangerous := range c.DangerousCommands {
		if strings.Contains(cmdLower, strings.ToLower(dangerous)) {
			return true
		}
	}
	
	// 其他安全检查规则
	// 检查是否包含管道重定向到特权文件
	if strings.Contains(cmdLower, "> /etc/") ||
	   strings.Contains(cmdLower, ">> /etc/") {
		return true
	}
	
	// 检查是否包含危险的shell操作
	if strings.Contains(cmdLower, ":(){ :|:& };:") { // fork炸弹
		return true
	}
	
	return false
}

// GetWarningMessage 获取警告信息
func (c *DefaultSecurityChecker) GetWarningMessage(command string) string {
	if !c.IsDangerousCommand(command) {
		return ""
	}
	
	var dangerType string
	cmdLower := strings.ToLower(command)
	
	switch {
	case strings.Contains(cmdLower, "rm"):
		dangerType = "文件删除"
	case strings.Contains(cmdLower, "chmod"):
		dangerType = "权限修改"
	case strings.Contains(cmdLower, "chown"):
		dangerType = "所有权修改"
	case strings.Contains(cmdLower, "mkfs"):
		dangerType = "文件系统格式化"
	case strings.Contains(cmdLower, "dd"):
		dangerType = "磁盘操作"
	case strings.Contains(cmdLower, "> /etc/") || strings.Contains(cmdLower, ">> /etc/"):
		dangerType = "系统配置修改"
	default:
		dangerType = "潜在危险操作"
	}
	
	return "警告：此命令包含" + dangerType + "操作，可能会导致数据丢失或系统问题。请确认您了解此命令的影响后再继续。"
} 