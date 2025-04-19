package deepseek

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/elecmonkey/prompt2cmd/internal/config"
	"github.com/elecmonkey/prompt2cmd/internal/history"
)

// Provider 实现DeepSeek API的LLM提供商
type Provider struct {
	APIKey  string
	BaseURL string
}

// NewProvider 创建一个新的DeepSeek API提供商
func NewProvider(config *config.Config) *Provider {
	return &Provider{
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	}
}

// IsLocal 返回是否为本地模型
func (p *Provider) IsLocal() bool {
	return false
}

// 用于构建多轮对话的消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExecutionAuditResult 命令执行审计结果
type ExecutionAuditResult struct {
	Success     bool   `json:"success"`      // 命令是否成功执行
	Description string `json:"description"`  // 对执行结果的解释
}

// GenerateCommand 根据提示和上下文生成命令和解释
func (p *Provider) GenerateCommand(prompt string, historyRecords []history.HistoryRecord) (string, string, error) {
	// 获取当前路径信息
	currentPath, err := os.Getwd()
	if err != nil {
		currentPath = "未知路径"
	}
	
	// 获取用户主目录，将绝对路径转换为~形式更简洁
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(currentPath, homeDir) {
		// 将用户主目录替换为~
		currentPath = filepath.Join("~", strings.TrimPrefix(currentPath, homeDir))
	}
	
	// 构建系统提示词，用于指导模型生成合适的命令
	systemPrompt := buildSystemPrompt(currentPath)
	
	// 创建消息数组，实现多轮对话
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加历史记录到消息数组中，构建对话历史
	// 为了实现类似于OpenAI文档中的多轮对话，我们需要交替添加用户和助手的消息
	for _, record := range historyRecords {
		// 添加用户的提示
		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: fmt.Sprintf("当前路径：%s\n用户需求：%s", currentPath, record.Prompt),
		})
		
		// 添加助手的回复（生成的命令）
		messages = append(messages, ChatMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("```\n%s\n```\n\n%s", record.Command, "已生成上述命令"),
		})
	}
	
	// 添加当前用户输入
	enhancedPrompt := fmt.Sprintf("当前路径：%s\n用户需求：%s", currentPath, prompt)
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: enhancedPrompt,
	})

	// 转换消息格式为map，以适应DeepSeek API的要求
	messagesMaps := make([]map[string]string, len(messages))
	for i, msg := range messages {
		messagesMaps[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// 创建请求体
	requestBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    messagesMaps,
		"temperature": 0.2, // 低温度以获得更确定性的响应
		"stream":      false,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	// 序列化请求体
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", errors.New("序列化请求失败: " + err.Error())
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", p.BaseURL+"/chat/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", "", errors.New("创建HTTP请求失败: " + err.Error())
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", errors.New("发送请求失败: " + err.Error())
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.New("读取响应失败: " + err.Error())
	}

	// 检查HTTP响应状态
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API调用失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", "", errors.New("解析响应失败: " + err.Error())
	}

	// 提取生成的内容
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", "", errors.New("未找到生成结果")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", "", errors.New("解析生成结果失败")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", "", errors.New("解析消息失败")
	}

	content, ok := message["content"].(string)
	if !ok || content == "" {
		return "", "", errors.New("生成内容为空")
	}

	// 解析JSON响应
	var parsedContent map[string]string
	err = json.Unmarshal([]byte(content), &parsedContent)
	if err != nil {
		return "", "", errors.New("解析JSON内容失败: " + err.Error())
	}

	// 提取命令和解释
	command, ok := parsedContent["command"]
	if !ok || command == "" {
		return "", "", errors.New("未找到生成的命令")
	}

	explanation, ok := parsedContent["explanation"]
	if !ok {
		explanation = "未提供命令解释"
	}

	return command, explanation, nil
}

// AuditExecutionResult 审计命令执行结果
func (p *Provider) AuditExecutionResult(command string, result string, prompt string) (*ExecutionAuditResult, error) {
	// 处理结果为空的情况
	resultContent := result
	if strings.TrimSpace(result) == "" {
		resultContent = "[无任何输出]"
	}
	
	// 构建系统提示词，用于指导模型审计命令执行结果
	systemPrompt := `你是一个命令执行结果审计专家。你需要根据执行结果判断命令是否成功执行。
请分析以下信息：
1. 用户的原始需求
2. 执行的命令
3. 命令的执行结果

你需要回答：
1. success：命令是否成功执行（布尔值true或false）
2. description：对执行结果的解释，说明为什么你认为命令成功或失败

请按照以下JSON格式返回：
{
  "success": true/false,
  "description": "解释命令执行结果的原因"
}

请注意：
- 如果命令正常执行但没有输出，通常也视为成功
- 如果结果中包含错误信息，通常表示命令失败
- 但有些命令执行结果中即使有"error"字样，也可能是正常输出的一部分
- 命令执行成功并不一定意味着满足了用户的需求，请根据用户需求和命令执行结果综合判断
- 无输出不一定意味着失败，某些命令执行成功后可能没有输出`

	// 创建消息数组
	messages := []map[string]string{
		{
			"role":    "system",
			"content": systemPrompt,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("用户需求: %s\n执行的命令: %s\n执行结果:\n%s", prompt, command, resultContent),
		},
	}

	// 创建请求体
	requestBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    messages,
		"temperature": 0.1, // 低温度以获得更确定性的响应
		"stream":      false,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	// 序列化请求体
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.New("序列化请求失败: " + err.Error())
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", p.BaseURL+"/chat/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, errors.New("创建HTTP请求失败: " + err.Error())
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("发送请求失败: " + err.Error())
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取响应失败: " + err.Error())
	}

	// 检查HTTP响应状态
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API调用失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, errors.New("解析响应失败: " + err.Error())
	}

	// 提取生成的内容
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, errors.New("未找到生成结果")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("解析生成结果失败")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, errors.New("解析消息失败")
	}

	content, ok := message["content"].(string)
	if !ok || content == "" {
		return nil, errors.New("生成内容为空")
	}

	// 解析JSON响应
	var auditResult ExecutionAuditResult
	err = json.Unmarshal([]byte(content), &auditResult)
	if err != nil {
		return nil, errors.New("解析JSON审计结果失败: " + err.Error())
	}

	return &auditResult, nil
}

// buildSystemPrompt 构建系统提示词
func buildSystemPrompt(currentPath string) string {
	osType := runtime.GOOS
	
	// 为不同操作系统提供更具体的示例
	osSpecificExamples := ""
	shellInfo := ""
	
	switch osType {
	case "windows":
		shellInfo = "默认使用PowerShell环境"
		osSpecificExamples = `
Windows系统命令示例（PowerShell）：
1. 列出目录内容：Get-ChildItem 或 ls
2. 查找文件：Get-ChildItem -Recurse -Filter "file.txt" 或 where.exe file.txt
3. 查看文件内容：Get-Content file.txt 或 cat file.txt
4. 删除文件：Remove-Item file.txt 或 del file.txt
5. 创建目录：New-Item -ItemType Directory -Name newdir 或 mkdir newdir
6. 查找文本：Select-String -Pattern "text" -Path file.txt 或 findstr "text" file.txt
7. 路径使用反斜杠或正斜杠：C:\\Users\\username\\Documents 或 C:/Users/username/Documents
8. 环境变量使用$前缀：$env:USERPROFILE
9. 管道操作使用 | 符号：Get-Process | Where-Object { $_.CPU -gt 10 }
10. 条件语句：if ($true) { "True" } else { "False" }
11. 循环：foreach ($item in $collection) { $item }
12. 常用别名：cd, ls, rm, mv, cp (这些别名使PowerShell命令与Linux/macOS命令兼容)`
	case "darwin":
		shellInfo = "默认使用Bash环境"
		osSpecificExamples = `
macOS系统命令示例：
1. 列出目录内容：ls -la
2. 查找文件：find . -name "file.txt" 或 mdfind "file.txt"
3. 查看文件内容：cat file.txt
4. 删除文件：rm file.txt
5. 创建目录：mkdir newdir
6. 查找文本：grep "text" file.txt
7. 路径使用正斜杠：/Users/username/Documents
8. 环境变量使用$前缀：$HOME
9. 管道操作使用 | 符号：ps aux | grep chrome
10. 条件语句：if [ $count -gt 0 ]; then echo "True"; else echo "False"; fi
11. 循环：for i in {1..5}; do echo $i; done
12. 权限管理：chmod 755 file.sh`
	default: // linux
		shellInfo = "默认使用Bash环境"
		osSpecificExamples = `
Linux系统命令示例：
1. 列出目录内容：ls -la
2. 查找文件：find . -name "file.txt" 或 locate "file.txt"
3. 查看文件内容：cat file.txt
4. 删除文件：rm file.txt
5. 创建目录：mkdir newdir
6. 查找文本：grep "text" file.txt
7. 路径使用正斜杠：/home/username/documents
8. 环境变量使用$前缀：$HOME
9. 管道操作使用 | 符号：ps aux | grep chrome
10. 条件语句：if [ $count -gt 0 ]; then echo "True"; else echo "False"; fi
11. 循环：for i in {1..5}; do echo $i; done
12. 权限管理：chmod 755 file.sh`
	}
	
	return fmt.Sprintf(`你是一个终端命令生成助手。你的任务是根据用户的自然语言描述，生成相应的终端命令。
当前操作系统：%s (%s)
当前工作路径：%s

请遵循以下规则：
1. 只生成与用户需求相关的命令
2. 提供命令的详细解释
3. 考虑当前工作路径，生成合适的命令
4. 务必生成适用于当前操作系统(%s)的命令，不要生成其他操作系统的命令
5. 要准确理解用户的真实意图，尤其是关于删除、修改等敏感操作
6. 输出必须是有效的JSON格式，包含以下字段：
   - command: 生成的终端命令
   - explanation: 命令的详细解释

%s

通用示例：
用户需求："列出当前目录下的所有图片文件"
{
  "command": "find . -type f -name \"*.jpg\" -o -name \"*.png\" -o -name \"*.gif\"",
  "explanation": "查找当前目录及其子目录下所有.jpg、.png和.gif格式的图片文件。"
}

用户需求："删除当前目录下所有.c文件"
{
  "command": "rm *.c",
  "explanation": "删除当前目录下所有以.c为扩展名的文件。"
}`, osType, shellInfo, currentPath, osType, osSpecificExamples)
} 