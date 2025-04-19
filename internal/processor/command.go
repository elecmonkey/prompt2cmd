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
}

// NewOSCommandProcessor 创建一个新的操作系统命令处理器
func NewOSCommandProcessor() *OSCommandProcessor {
	return &OSCommandProcessor{
		Platform: runtime.GOOS,
	}
}

// ProcessCommand 根据操作系统处理命令
func (p *OSCommandProcessor) ProcessCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("命令不能为空")
	}

	// 根据操作系统调整命令
	switch p.Platform {
	case "windows":
		// Windows系统中，一些命令需要特别处理
		if strings.HasPrefix(command, "ls") {
			command = strings.Replace(command, "ls", "dir", 1)
		}
		// 可以添加更多Windows特定的命令转换
	case "darwin", "linux":
		// Unix类系统命令通常兼容
		// 但有些特殊情况也需处理
	}

	return command, nil
}

// ExecuteCommand 执行命令并返回结果
func (p *OSCommandProcessor) ExecuteCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("命令不能为空")
	}

	// 根据平台选择合适的shell
	var cmd *exec.Cmd
	if p.Platform == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

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