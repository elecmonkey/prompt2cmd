# Prompt2Cmd

Prompt2Cmd是一个终端工具，能够通过自然语言理解用户需求，生成相应的命令，并在用户确认后执行。该工具利用大型语言模型(LLM)实现自然语言到命令的转换，提高用户使用终端的效率和体验。

## 特性

- **自然语言输入**：用日常语言描述想要执行的操作
- **命令生成**：利用LLM将自然语言转换为终端命令
- **命令解释**：提供命令的详细说明
- **安全检查**：对危险命令(如rm, chmod等)添加额外警告
- **多平台支持**：根据操作系统自动调整命令(Linux/macOS)
- **命令历史记录**：保存生成和执行过的命令
- **上下文感知**：使用最近5条命令历史作为上下文，支持连续对话
- **直接执行cd命令**：对于"cd "开头的指令直接执行，无需通过LLM
- **显示当前路径**：在提示符中显示当前工作路径
- **执行结果审计**：使用LLM评估命令执行结果，判断是否成功完成用户需求，包括错误分析
- **灵活配置管理**：支持从多个位置自动查找配置文件

## 安装

### 前提条件

- Go 1.16+
- DeepSeek API密钥（需在[DeepSeek平台](https://platform.deepseek.com/api_keys)申请）

### 安装步骤

1. 克隆仓库：

```bash
git clone https://github.com/elecmonkey/prompt2cmd.git
cd prompt2cmd
```

2. 创建配置文件：

程序会按照以下顺序查找配置文件：
- 当前工作目录下的`.env`
- 用户主目录下的`~/.prompt2cmd/.env`
- 用户主目录下的`~/.prompt2cmd_env`
- 系统配置目录下的`/etc/prompt2cmd/.env`

选择其中一个位置创建`.env`文件，内容如下：

```
# DeepSeek API密钥 
DEEPSEEK_API_KEY=your_api_key_here

# API基础URL
DEEPSEEK_BASE_URL=https://api.deepseek.com

# 应用配置
MAX_HISTORY_SIZE=50
USE_LOCAL_MODEL=false

# 安全设置
# 危险命令列表，使用逗号分隔
DANGEROUS_COMMANDS=rm -rf,rm,chmod,chown,mkfs,dd,mv,reboot,shutdown
```

3. 安装依赖并编译：

```bash
go mod tidy
go build -o prompt2cmd ./cmd/prompt2cmd
```

4. 将可执行文件移动到PATH路径（可选）：

```bash
# Linux/macOS
sudo mv prompt2cmd /usr/local/bin/

# 或将其保留在当前目录中运行
./prompt2cmd
```

> **注意**：如果你将程序移动到系统路径，建议将配置文件放在用户主目录下的`~/.prompt2cmd/.env`，这样可以确保程序在任何位置运行时都能找到配置。

## 使用方法

1. 启动程序：

```bash
prompt2cmd
```

2. 用自然语言描述你想执行的操作，例如：

```
🤖 (~/projects)你想要：查看当前目录下的所有图片文件，并按大小排序
```

3. 程序会生成相应的命令和解释：

```
📝 生成的命令:
   find . -type f -name "*.jpg" -o -name "*.png" -o -name "*.gif" | xargs ls -lhS

📋 命令解释:
   查找当前目录及子目录下所有.jpg、.png和.gif格式的图片文件，然后使用ls命令按文件大小降序排列显示详细信息。
```

4. 确认、修改或取消命令：

```
❓ 是否执行此命令? (y/n/e[编辑]): 
```

5. 查看执行结果和审计结果：

```
🔍 执行结果:
--------------------------------------------------
-rw-r--r-- 1 user user 2.5M Apr 20 12:33 ./photos/vacation.jpg
-rw-r--r-- 1 user user 1.2M Apr 20 12:32 ./images/logo.png
-rw-r--r-- 1 user user 500K Apr 20 12:30 ./icons/button.gif
--------------------------------------------------

🔍 正在审计执行结果...
✅ 执行状态: true
📋 审计结果: 命令成功执行并满足了用户需求。命令找到了当前目录及子目录下的所有图片文件(.jpg, .png, .gif)并按照文件大小进行了排序显示，从大到小依次是vacation.jpg(2.5M)、logo.png(1.2M)和button.gif(500K)。
```

6. 当命令执行失败时，也会进行错误分析：

```
⚙️ 正在执行命令...
❌ 错误: exit status 1

🔍 正在审计执行结果...
❌ 执行状态: false
📋 审计结果: 命令执行失败。"find"命令的参数可能有误，或者目标路径不存在。错误代码"exit status 1"表示命令运行时出现了错误。建议检查命令语法或尝试简化命令，分步骤执行以确定具体问题。
```

7. 直接使用cd命令改变工作目录：

```
🤖 (~/projects)你想要：cd ~/documents
✅ 已切换到目录: /home/user/documents
```

## 配置说明

### 配置文件位置

程序会按以下顺序查找配置文件：

1. 如果指定了配置文件路径（通过参数），优先使用该路径
2. 当前工作目录下的`.env`
3. 用户主目录下的`~/.prompt2cmd/.env`
4. 用户主目录下的`~/.prompt2cmd_env`（兼容性考虑）
5. 系统配置目录下的`/etc/prompt2cmd/.env`

### 配置项说明

| 配置项 | 描述 | 必需 | 默认值 |
|-------|------|------|-------|
| DEEPSEEK_API_KEY | DeepSeek API密钥 | 是 | 无 |
| DEEPSEEK_BASE_URL | DeepSeek API基础URL | 否 | https://api.deepseek.com |
| MAX_HISTORY_SIZE | 历史记录最大保存数量 | 否 | 50 |
| USE_LOCAL_MODEL | 是否使用本地模型 | 否 | false |
| LOCAL_MODEL_PATH | 本地模型路径 | 仅当USE_LOCAL_MODEL=true时必需 | 无 |
| DANGEROUS_COMMANDS | 危险命令列表（逗号分隔） | 否 | rm -rf,rm,chmod,chown,mkfs,dd,mv,reboot,shutdown |

## 安全注意事项

- 所有命令在执行前都需要用户确认
- 危险命令会有额外警告提示
- 建议在非关键环境中使用此工具

## 开发计划

- [ ] 支持本地模型
- [ ] 改进TUI界面
- [ ] 添加命令建议功能
- [ ] 支持更多的LLM提供商
- [ ] 实现插件系统

## 贡献

欢迎贡献代码、提交问题或改进建议！

## 许可证

MIT 