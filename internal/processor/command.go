package processor

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// CommandProcessor 命令处理器接口
type CommandProcessor interface {
	// ProcessCommand 处理命令，返回处理后的命令和可能的错误
	ProcessCommand(command string) (string, error)

	// ExecuteCommand 执行命令，返回执行结果和可能的错误
	ExecuteCommand(command string) (string, error)
}

// OSCommandProcessor 操作系统命令处理器
type OSCommandProcessor struct {
	Platform string // windows, linux, darwin
	UsePS    bool   // Windows下是否使用PowerShell
}

// NewOSCommandProcessor 创建一个新的操作系统命令处理器
func NewOSCommandProcessor() *OSCommandProcessor {
	processor := &OSCommandProcessor{
		Platform: runtime.GOOS,
	}

	// Windows系统默认使用PowerShell
	// if processor.Platform == "windows" {
	// 	processor.UsePS = true
	// }

	return processor
}

// ProcessCommand 根据操作系统处理命令
func (p *OSCommandProcessor) ProcessCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("命令不能为空")
	}

	// 命令处理基本上交给LLM完成，这里只做简单调整
	return command, nil
}

// ExecuteCommand 执行命令并返回结果
func (p *OSCommandProcessor) ExecuteCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("命令不能为空")
	}

	// 根据平台选择合适的shell
	var cmd *exec.Cmd
	// if p.Platform == "windows" {
	// 	if p.UsePS {
	// 		// 使用PowerShell
	// 		cmd = exec.Command("powershell", "-Command", command)
	// 	} else {
	// 		// 使用CMD（传统方式，保留以便兼容）
	// 		cmd = exec.Command("cmd", "/C", command)
	// 	}
	// } else {
	// Linux/macOS 使用标准shell
	cmd = exec.Command("sh", "-c", command)
	// }

	// 设置命令的输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// IsCommandSafe 检查命令是否安全（将在安全模块实现更详细的检查）
func IsCommandSafe(command string, dangerousCommands []string) bool {
	for _, dangerous := range dangerousCommands {
		if strings.Contains(command, dangerous) {
			return false
		}
	}
	return true
}
