package ui

import (
	"errors"
)

// UserInterface 用户界面接口
type UserInterface interface {
	// GetUserInput 获取用户输入
	GetUserInput() (string, error)
	
	// DisplayGeneratedCommand 显示生成的命令和解释
	DisplayGeneratedCommand(command string, explanation string)
	
	// GetUserConfirmation 获取用户确认
	GetUserConfirmation() (bool, error)
	
	// DisplayExecutionResult 显示执行结果
	DisplayExecutionResult(result string)
	
	// DisplayError 显示错误信息
	DisplayError(err error)
}

// InputValidator 输入验证器
type InputValidator interface {
	ValidateInput(input string) error
}

// DefaultInputValidator 默认输入验证器
type DefaultInputValidator struct{}

// ValidateInput 验证用户输入
func (v *DefaultInputValidator) ValidateInput(input string) error {
	if input == "" {
		return errors.New("输入不能为空")
	}
	return nil
} 