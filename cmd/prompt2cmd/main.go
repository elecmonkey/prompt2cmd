package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elecmonkey/prompt2cmd/internal/config"
	"github.com/elecmonkey/prompt2cmd/internal/history"
	"github.com/elecmonkey/prompt2cmd/internal/llm/deepseek"
	"github.com/elecmonkey/prompt2cmd/internal/processor"
	"github.com/elecmonkey/prompt2cmd/internal/security"
	"github.com/elecmonkey/prompt2cmd/internal/ui"
)

const (
	appVersion = "0.1.0"
	envFile    = ".env"
)

func main() {
	fmt.Printf("🚀 Prompt2Cmd v%s - 自然语言转终端命令工具\n", appVersion)
	fmt.Println("输入 'exit' 或 'quit' 退出程序")

	// 获取当前工作目录
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取当前工作目录失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 初始化配置管理器
	configPath := filepath.Join(workingDir, envFile)
	configManager := config.NewEnvConfigManager(configPath)
	cfg, err := configManager.LoadConfig()
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 初始化 LLM 提供商
	llmProvider := deepseek.NewProvider(cfg)

	// 初始化命令处理器
	cmdProcessor := processor.NewOSCommandProcessor()

	// 初始化安全检查器
	securityChecker := security.NewSecurityChecker(cfg.DangerousCommands)

	// 初始化历史记录
	historyManager, err := history.NewFileCommandHistory("", cfg.MaxHistorySize)
	if err != nil {
		fmt.Printf("⚠️ 历史记录功能不可用: %s\n", err.Error())
		fmt.Println("程序将继续运行，但不会记录命令历史")
		// 创建一个临时的内存历史记录管理器
		historyManager = &history.FileCommandHistory{
			MaxRecords: cfg.MaxHistorySize,
		}
	}

	// 初始化用户界面
	userInterface := ui.NewTerminalUI()

	// 启动主循环
	reader := bufio.NewReader(os.Stdin)
	// 使用最近5条历史记录
	contextLimit := 5 // 上下文记录数量

	for {
		// 获取用户输入
		prompt, err := userInterface.GetUserInput()
		if err != nil {
			userInterface.DisplayError(err)
			continue
		}

		// 检查退出命令
		if prompt == "exit" || prompt == "quit" || prompt == "退出" {
			fmt.Println("👋 再见!")
			break
		}

		// 检查是否为cd命令
		if strings.HasPrefix(prompt, "cd ") {
			// 直接处理cd命令
			dirPath := strings.TrimSpace(strings.TrimPrefix(prompt, "cd "))
			
			// 处理特殊情况：~表示用户主目录
			if strings.HasPrefix(dirPath, "~") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					dirPath = filepath.Join(homeDir, strings.TrimPrefix(dirPath, "~"))
				}
			}
			
			// 改变工作目录
			err := os.Chdir(dirPath)
			if err != nil {
				userInterface.DisplayError(fmt.Errorf("切换目录失败: %s", err.Error()))
			} else {
				currentDir, _ := os.Getwd()
				fmt.Printf("\n✅ 已切换到目录: %s\n", currentDir)
				
				// 添加到历史记录
				_ = historyManager.AddCommand(prompt, prompt, true)
			}
			continue
		}

		// 获取历史记录
		historyRecords, err := historyManager.GetHistory(contextLimit)
		if err != nil {
			fmt.Printf("⚠️ 无法获取历史记录: %s\n", err.Error())
			fmt.Println("将继续生成命令，但不使用历史上下文")
			historyRecords = []history.HistoryRecord{}
		}

		// 生成命令
		fmt.Println("\n🔄 正在生成命令...")
		// 使用历史记录作为上下文
		command, explanation, err := llmProvider.GenerateCommand(prompt, historyRecords)
		if err != nil {
			userInterface.DisplayError(err)
			continue
		}

		// 根据操作系统处理命令
		command, err = cmdProcessor.ProcessCommand(command)
		if err != nil {
			userInterface.DisplayError(err)
			continue
		}

		// 显示生成的命令和解释
		userInterface.DisplayGeneratedCommand(command, explanation)

		// 检查命令安全性
		if securityChecker.IsDangerousCommand(command) {
			fmt.Printf("\n⚠️ %s\n", securityChecker.GetWarningMessage(command))
		}

		// 获取用户确认
	confirmLoop:
		for {
			confirmed, err := userInterface.GetUserConfirmation()
			if err != nil {
				if err.Error() == "EDIT_COMMAND" {
					// 用户要求编辑命令
					fmt.Print("\n✏️ 请编辑命令: ")
					command, err = reader.ReadString('\n')
					if err != nil {
						userInterface.DisplayError(err)
						break confirmLoop
					}
					command = strings.TrimSpace(command)
					if command == "" {
						userInterface.DisplayError(fmt.Errorf("命令不能为空"))
						break confirmLoop
					}
					continue
				}
				userInterface.DisplayError(err)
				break
			}

			if !confirmed {
				fmt.Println("\n❌ 命令已取消")
				break
			}

			// 执行命令
			fmt.Println("\n⚙️ 正在执行命令...")
			result, execErr := cmdProcessor.ExecuteCommand(command)
			
			// 显示执行结果（无论成功还是失败）
			if execErr != nil {
				userInterface.DisplayError(execErr)
				result = fmt.Sprintf("执行失败: %s", execErr.Error())
			} else {
				userInterface.DisplayExecutionResult(result)
			}

			// 使用LLM审计执行结果
			fmt.Println("\n🔍 正在审计执行结果...")
			auditResult, err := llmProvider.AuditExecutionResult(command, result, prompt)
			if err != nil {
				fmt.Printf("❌ 审计失败: %s\n", err.Error())
			} else {
				// 显示审计结果
				statusEmoji := "✅"
				if !auditResult.Success {
					statusEmoji = "❌"
				}
				fmt.Printf("\n%s 执行状态: %v\n", statusEmoji, auditResult.Success)
				fmt.Printf("📋 审计结果: %s\n", auditResult.Description)
			}

			// 添加到历史记录（根据是否有执行错误决定成功状态）
			_ = historyManager.AddCommand(prompt, command, execErr == nil)
			break
		}
	}
} 