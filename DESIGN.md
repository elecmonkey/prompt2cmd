# Prompt2Cmd 设计文档

## 1. 项目概述

Prompt2Cmd是一个终端工具，能够通过自然语言理解用户需求，生成相应的命令，并在用户确认后执行。该工具利用大型语言模型(LLM)实现自然语言到命令的转换，提高用户使用终端的效率和体验。

## 2. 核心功能

### 2.1 基础功能

- **自然语言输入**：用户用日常语言描述想要执行的操作
- **命令生成**：LLM将自然语言转换为终端命令
- **命令解释**：同时提供命令的详细说明
- **用户确认**：显示生成的命令并要求用户确认
- **执行或修改**：用户可以选择执行、修改或放弃命令

### 2.2 增强功能

- **命令历史记录**：保存生成和执行过的命令
- **多平台支持**：根据操作系统自动调整命令(Windows/Linux/macOS)
- **安全检查**：对危险命令(如rm, chmod等)添加额外警告
- **上下文感知**：记住之前的命令，支持后续命令的上下文
- **本地模型支持**：可选项使用本地运行的LLM以减少API依赖

## 3. 系统架构

### 3.1 整体架构

Prompt2Cmd采用模块化设计，主要包含以下模块：

1. **用户交互模块**：处理用户输入和输出
2. **LLM接口模块**：与语言模型API通信
3. **命令处理模块**：处理命令的生成、解释和执行
4. **安全检查模块**：对命令进行安全性检查
5. **历史记录模块**：管理命令历史
6. **配置管理模块**：管理用户配置和系统设置

### 3.2 数据流

```
用户输入 -> 用户交互模块 -> LLM接口模块 -> 命令处理模块 -> 安全检查模块 -> 用户确认 -> 命令执行 -> 历史记录
```

## 4. 技术栈选择

- **编程语言**：Go
- **LLM API**：DeepSeek API（主要）
- **本地模型**：支持如Ollama等本地部署的模型
- **UI框架**：使用bubbletea/charm等TUI库实现终端界面
- **存储**：命令历史和配置使用本地文件存储(如JSON/YAML)

## 5. 模块设计

### 5.1 用户交互模块

```go
type UserInterface interface {
    GetUserInput() (string, error)
    DisplayGeneratedCommand(command string, explanation string)
    GetUserConfirmation() (bool, error)
    DisplayExecutionResult(result string)
    DisplayError(err error)
}
```

### 5.2 LLM接口模块

```go
type LLMProvider interface {
    GenerateCommand(prompt string, context []string) (string, string, error) // 返回命令和解释
    IsLocal() bool
}

type DeepSeekProvider struct {
    ApiKey string
    BaseURL string
}

type LocalModelProvider struct {
    ModelPath string
    // 本地模型相关配置
}
```

### 5.3 命令处理模块

```go
type CommandProcessor interface {
    ProcessCommand(command string) (string, error)
    ExecuteCommand(command string) (string, error)
}

type OSCommandProcessor struct {
    Platform string // windows, linux, darwin
}
```

### 5.4 安全检查模块

```go
type SecurityChecker interface {
    IsDangerousCommand(command string) bool
    GetWarningMessage(command string) string
}
```

### 5.5 历史记录模块

```go
type CommandHistory interface {
    AddCommand(prompt string, command string, executed bool)
    GetHistory(limit int) []HistoryRecord
    GetContextForPrompt(limit int) []string
}
```

### 5.6 配置管理模块

```go
type ConfigManager interface {
    LoadConfig() (Config, error)
    SaveConfig(Config) error
    GetLLMProvider() LLMProvider
}

type Config struct {
    APIKey string
    UseLocalModel bool
    LocalModelPath string
    MaxHistorySize int
    DangerousCommands []string
}
```

## 6. 安全考虑

### 6.1 命令执行安全

- 所有命令必须经过用户确认后才能执行
- 危险命令列表配置，检测到危险命令时提供额外警告
- 支持命令白名单/黑名单配置

### 6.2 API密钥安全

- API密钥加密存储
- 支持环境变量或配置文件方式提供密钥

### 6.3 数据隐私

- 本地历史记录加密存储
- 可选择仅使用本地模型，减少数据传输

## 7. 用户体验设计

### 7.1 交互流程

1. 用户输入自然语言需求
2. 系统生成命令和解释
3. 显示命令并请求确认
4. 用户选择执行/修改/放弃
5. 执行命令并显示结果

### 7.2 命令展示方式

- 命令高亮显示
- 命令解释使用不同颜色
- 危险命令警告突出显示

### 7.3 错误处理

- 友好的错误提示
- 对常见错误提供解决建议
- 命令执行失败时提供反馈

## 8. 实现路线图

### 8.1 第一阶段（核心功能）

- 实现基础用户交互
- 接入DeepSeek API
- 实现命令生成和执行
- 简单的安全检查

### 8.2 第二阶段（增强功能）

- 命令历史记录
- 多平台支持优化
- 完善安全检查
- 上下文感知功能

### 8.3 第三阶段（高级功能）

- 本地模型支持
- TUI界面优化
- 插件系统（可扩展）
- 自定义提示词模板

## 9. 目录结构

```
prompt2cmd/
├── cmd/
│   └── prompt2cmd/             # 主入口
├── internal/
│   ├── config/                 # 配置管理
│   ├── history/                # 历史记录
│   ├── llm/                    # LLM接口
│   │   ├── deepseek/           # DeepSeek API实现
│   │   └── local/              # 本地模型实现
│   ├── processor/              # 命令处理
│   ├── security/               # 安全检查
│   └── ui/                     # 用户界面
├── pkg/                        # 可重用包
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 10. 后续计划

- 支持更多的LLM提供商
- 多语言支持
- 命令模板系统
- 命令修改建议功能
- Web界面（可选）

## 11. 总结

Prompt2Cmd是一个强大的终端工具，通过LLM技术实现自然语言到命令的转换，同时注重安全性和用户体验。采用Go语言实现，结合模块化设计，使系统具有良好的可维护性和扩展性。通过持续改进和功能迭代，该工具将成为提高终端使用效率的有力助手。 