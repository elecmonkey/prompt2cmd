package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 存储应用程序配置
type Config struct {
	LLMProvider       string // deepseek, moonshot
	LLMAPIKey         string
	LLMBaseURL        string
	LLMModel          string
	UseLocalModel     bool
	LocalModelPath    string
	MaxHistorySize    int
	DangerousCommands []string
	// 添加一个配置文件路径，以便后续可能的配置保存
	ConfigFile string
}

// ConfigManager 接口定义配置管理器的行为
type ConfigManager interface {
	LoadConfig() (*Config, error)
}

// EnvConfigManager 从环境变量加载配置
type EnvConfigManager struct {
	envFile string
}

// NewEnvConfigManager 创建一个新的环境变量配置管理器
func NewEnvConfigManager(envFile string) *EnvConfigManager {
	return &EnvConfigManager{
		envFile: envFile,
	}
}

// 寻找配置文件的可能位置
func findConfigFile(envFile string) (string, error) {
	// 如果提供了明确的配置文件路径，直接使用
	if envFile != "" {
		if _, err := os.Stat(envFile); err == nil {
			fmt.Printf("使用指定的配置文件: %s\n", envFile)
			return envFile, nil
		}
	}

	// 定义可能的配置文件位置
	var configPaths []string

	// 1. 当前工作目录
	currentDir, err := os.Getwd()
	if err == nil {
		configPaths = append(configPaths, filepath.Join(currentDir, ".env"))
	}

	// 2. 用户主目录
	homeDir, err := os.UserHomeDir()
	if err == nil {
		// ~/.prompt2cmd/.env
		promptConfigDir := filepath.Join(homeDir, ".prompt2cmd")
		configPaths = append(configPaths, filepath.Join(promptConfigDir, ".env"))
		// 直接放在~目录下的.prompt2cmd_env（兼容性考虑）
		configPaths = append(configPaths, filepath.Join(homeDir, ".prompt2cmd_env"))
	}

	// 3. 系统配置目录
	configPaths = append(configPaths, "/etc/prompt2cmd/.env")

	// 查找第一个存在的配置文件
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("找到配置文件: %s\n", path)
			return path, nil
		}
	}

	// 如果未找到任何配置文件，返回空字符串和错误
	return "", errors.New("未找到配置文件。请在当前目录、用户主目录或系统配置目录创建.env文件")
}

// 确保配置目录存在
func ensureConfigDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// LoadConfig 从.env文件加载配置
func (e *EnvConfigManager) LoadConfig() (*Config, error) {
	// 创建配置对象
	config := &Config{}

	// 寻找配置文件
	configFile, err := findConfigFile(e.envFile)
	if err == nil {
		// 加载找到的.env文件
		err = godotenv.Load(configFile)
		if err != nil {
			fmt.Printf("警告: 加载配置文件失败: %s\n", err.Error())
		} else {
			// 保存找到的配置文件路径
			config.ConfigFile = configFile
		}
	} else {
		fmt.Printf("警告: %s\n", err.Error())
		fmt.Println("将使用环境变量或默认值...")
	}

	// 获取LLM提供商（必需）
	config.LLMProvider = os.Getenv("LLM_PROVIDER")
	if config.LLMProvider == "" {
		config.LLMProvider = "deepseek" // 默认使用deepseek
	}

	// 获取LLM API密钥（必需）
	config.LLMAPIKey = os.Getenv("LLM_API_KEY")
	if config.LLMAPIKey == "" {
		// 尝试创建用户配置目录和示例配置
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			promptConfigDir := filepath.Join(homeDir, ".prompt2cmd")
			if err := ensureConfigDir(promptConfigDir); err == nil {
				exampleConfigPath := filepath.Join(promptConfigDir, ".env.example")
				if _, err := os.Stat(exampleConfigPath); os.IsNotExist(err) {
					// 创建示例配置文件
					exampleConfig := `# Prompt2Cmd 配置文件示例
# 在 https://platform.deepseek.com/api_keys 或 https://platform.moonshot.cn/console/api-keys 获取API密钥

# LLM 提供商 (deepseek, moonshot, 默认为 deepseek)
LLM_PROVIDER=deepseek

# LLM API 密钥（必需）
LLM_API_KEY=your_api_key_here

# LLM API 基础URL（可选，有默认值）
# DeepSeek: https://api.deepseek.com
# Moonshot: https://api.moonshot.cn/v1
LLM_BASE_URL=

# LLM 模型名称（可选，有默认值）
# DeepSeek: deepseek-chat
# Moonshot: kimi-k2-0711-preview
LLM_MODEL=

# 历史记录最大保存数量（可选，有默认值）
MAX_HISTORY_SIZE=50

# 是否使用本地模型（可选，有默认值）
USE_LOCAL_MODEL=false

# 本地模型路径（仅当USE_LOCAL_MODEL=true时必需）
LOCAL_MODEL_PATH=

# 危险命令列表（可选，有默认值）
DANGEROUS_COMMANDS=rm -rf,rm,chmod,chown,mkfs,dd,mv,reboot,shutdown
`
					err := os.WriteFile(exampleConfigPath, []byte(exampleConfig), 0644)
					if err == nil {
						fmt.Printf("已在 %s 创建示例配置文件\n", exampleConfigPath)
					}
				}
			}
		}

		return nil, errors.New("未找到LLM_API_KEY环境变量，这是必需的。请设置环境变量或在配置文件中提供")
	}

	// 获取LLM基础URL
	config.LLMBaseURL = os.Getenv("LLM_BASE_URL")
	if config.LLMBaseURL == "" {
		// 根据提供商设置默认URL
		switch config.LLMProvider {
		case "deepseek":
			config.LLMBaseURL = "https://api.deepseek.com"
		case "moonshot":
			config.LLMBaseURL = "https://api.moonshot.cn/v1"
		default:
			return nil, errors.New("不支持的LLM提供商: " + config.LLMProvider)
		}
	}

	// 获取LLM模型名称
	config.LLMModel = os.Getenv("LLM_MODEL")
	if config.LLMModel == "" {
		// 根据提供商设置默认模型
		switch config.LLMProvider {
		case "deepseek":
			config.LLMModel = "deepseek-chat"
		case "moonshot":
			config.LLMModel = "kimi-k2-0711-preview"
		default:
			return nil, errors.New("不支持的LLM提供商: " + config.LLMProvider)
		}
	}

	// 获取是否使用本地模型
	useLocalModelStr := os.Getenv("USE_LOCAL_MODEL")
	if useLocalModelStr != "" {
		config.UseLocalModel = strings.ToLower(useLocalModelStr) == "true"
	} else {
		config.UseLocalModel = false // 默认不使用本地模型
	}

	// 获取本地模型路径
	config.LocalModelPath = os.Getenv("LOCAL_MODEL_PATH")
	// 如果设置了使用本地模型但没有提供路径，返回错误
	if config.UseLocalModel && config.LocalModelPath == "" {
		return nil, errors.New("启用了本地模型(USE_LOCAL_MODEL=true)，但未设置LOCAL_MODEL_PATH")
	}

	// 获取历史记录大小限制
	config.MaxHistorySize = 50 // 默认值
	maxHistorySizeStr := os.Getenv("MAX_HISTORY_SIZE")
	if maxHistorySizeStr != "" {
		maxHistorySize, err := strconv.Atoi(maxHistorySizeStr)
		if err != nil {
			return nil, errors.New("MAX_HISTORY_SIZE必须是一个有效的整数: " + err.Error())
		}
		if maxHistorySize < 1 {
			return nil, errors.New("MAX_HISTORY_SIZE必须大于0")
		}
		config.MaxHistorySize = maxHistorySize
	}

	// 获取危险命令列表
	config.DangerousCommands = []string{"rm -rf", "rm", "chmod", "chown", "mkfs", "dd", "mv", "reboot", "shutdown"} // 默认列表
	dangerousCommandsStr := os.Getenv("DANGEROUS_COMMANDS")
	if dangerousCommandsStr != "" {
		// 分割字符串并清理空格
		dangerousCommands := strings.Split(dangerousCommandsStr, ",")
		for i, cmd := range dangerousCommands {
			dangerousCommands[i] = strings.TrimSpace(cmd)
		}
		// 过滤空字符串
		var filteredCommands []string
		for _, cmd := range dangerousCommands {
			if cmd != "" {
				filteredCommands = append(filteredCommands, cmd)
			}
		}
		if len(filteredCommands) > 0 {
			config.DangerousCommands = filteredCommands
		}
	}

	return config, nil
}
