package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TerminalUI 终端用户界面
type TerminalUI struct {
	validator InputValidator
	reader    *bufio.Reader
}

// NewTerminalUI 创建一个新的终端用户界面
func NewTerminalUI() *TerminalUI {
	return &TerminalUI{
		validator: &DefaultInputValidator{},
		reader:    bufio.NewReader(os.Stdin),
	}
}

// GetUserInput 获取用户输入
func (ui *TerminalUI) GetUserInput() (string, error) {
	// 获取当前路径
	currentPath, err := os.Getwd()
	if err != nil {
		currentPath = "未知路径"
	}
	
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(currentPath, homeDir) {
		// 将用户主目录替换为~
		currentPath = filepath.Join("~", strings.TrimPrefix(currentPath, homeDir))
	}
	
	fmt.Printf("\n🤖 (%s) 你想要：", currentPath)
	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return "", errors.New("读取输入失败: " + err.Error())
	}
	
	// 清理输入
	input = strings.TrimSpace(input)
	
	// 验证输入
	err = ui.validator.ValidateInput(input)
	if err != nil {
		return "", err
	}
	
	return input, nil
}

// DisplayGeneratedCommand 显示生成的命令和解释
func (ui *TerminalUI) DisplayGeneratedCommand(command string, explanation string) {
	fmt.Println("\n📝 生成的命令:")
	fmt.Printf("   \033[1;36m%s\033[0m\n", command)
	fmt.Println("\n📋 命令解释:")
	fmt.Printf("   %s\n", explanation)
}

// GetUserConfirmation 获取用户确认
func (ui *TerminalUI) GetUserConfirmation() (bool, error) {
	fmt.Print("\n❓ 是否执行此命令? (y/n/e[编辑]): ")
	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return false, errors.New("读取输入失败: " + err.Error())
	}
	
	input = strings.TrimSpace(strings.ToLower(input))
	
	switch input {
	case "y", "yes", "是":
		return true, nil
	case "n", "no", "否":
		return false, nil
	case "e", "edit", "编辑":
		// 返回特殊错误，表示用户想编辑命令
		return false, errors.New("EDIT_COMMAND")
	default:
		return false, errors.New("无效输入，请输入 y/n/e")
	}
}

// DisplayExecutionResult 显示执行结果
func (ui *TerminalUI) DisplayExecutionResult(result string) {
	fmt.Println("\n🔍 执行结果:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(result)
	fmt.Println("--------------------------------------------------")
}

// DisplayError 显示错误信息
func (ui *TerminalUI) DisplayError(err error) {
	if err == nil {
		return
	}
	
	fmt.Printf("\n❌ 错误: %s\n", err.Error())
} 