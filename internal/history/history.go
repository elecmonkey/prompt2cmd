package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HistoryRecord 表示一条命令历史记录
type HistoryRecord struct {
	ID        string `json:"id"`
	Prompt    string `json:"prompt"`
	Command   string `json:"command"`
	Executed  bool   `json:"executed"`
	Timestamp string `json:"timestamp"`
}

// CommandHistory 接口定义了命令历史记录的行为
type CommandHistory interface {
	AddCommand(prompt, command string, executed bool) error
	GetHistory(limit int) ([]HistoryRecord, error)
}

// FileCommandHistory 使用文件存储命令历史记录
type FileCommandHistory struct {
	filePath  string
	records   []HistoryRecord
	MaxRecords int
}

// 查找历史文件的可能位置
func findHistoryFile(filePath string) (string, error) {
	// 如果提供了明确的文件路径，直接使用
	if filePath != "" {
		return filePath, nil
	}

	// 定义可能的历史文件位置
	var historyPaths []string

	// 1. 用户主目录
	homeDir, err := os.UserHomeDir()
	if err == nil {
		// ~/.prompt2cmd/history.json
		promptDir := filepath.Join(homeDir, ".prompt2cmd")
		historyPaths = append(historyPaths, filepath.Join(promptDir, "history.json"))
		// 直接放在~目录下的.prompt2cmd_history（兼容性考虑）
		historyPaths = append(historyPaths, filepath.Join(homeDir, ".prompt2cmd_history"))
	}

	// 先尝试查找已存在的文件
	for _, path := range historyPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// 如果未找到，则使用第一个路径（但确保目录存在）
	if len(historyPaths) > 0 {
		firstChoice := historyPaths[0]
		// 确保目录存在
		dir := filepath.Dir(firstChoice)
		if err := os.MkdirAll(dir, 0755); err == nil {
			return firstChoice, nil
		}
	}

	// 如果都失败了，返回默认值
	if homeDir != "" {
		return filepath.Join(homeDir, ".prompt2cmd_history"), nil
	}
	
	return "", errors.New("无法确定历史记录文件路径")
}

// NewFileCommandHistory 创建一个新的文件命令历史记录
func NewFileCommandHistory(filePath string, maxRecords int) (*FileCommandHistory, error) {
	if maxRecords <= 0 {
		maxRecords = 50 // 默认值，防止无效值
	}

	// 查找或创建历史文件
	resolvedPath, err := findHistoryFile(filePath)
	if err != nil {
		return nil, err
	}

	history := &FileCommandHistory{
		filePath: resolvedPath,
		MaxRecords: maxRecords,
	}

	// 尝试从文件加载记录
	if _, err := os.Stat(resolvedPath); err == nil {
		// 文件存在，尝试加载
		file, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, errors.New("读取历史记录文件失败: " + err.Error())
		}

		// 解析JSON
		if len(file) > 0 {
			err = json.Unmarshal(file, &history.records)
			if err != nil {
				return nil, errors.New("解析历史记录失败: " + err.Error())
			}
		}
	}

	return history, nil
}

// AddCommand 添加一条命令到历史记录
func (h *FileCommandHistory) AddCommand(prompt, command string, executed bool) error {
	// 创建新的历史记录
	record := HistoryRecord{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Prompt:    prompt,
		Command:   command,
		Executed:  executed,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 添加到记录数组
	h.records = append(h.records, record)

	// 确保记录数量不超过最大限制
	if len(h.records) > h.MaxRecords {
		h.records = h.records[len(h.records)-h.MaxRecords:]
	}

	// 序列化记录
	data, err := json.MarshalIndent(h.records, "", "  ")
	if err != nil {
		return errors.New("序列化历史记录失败: " + err.Error())
	}

	// 确保目录存在
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.New("创建历史记录目录失败: " + err.Error())
	}

	// 写入文件
	err = os.WriteFile(h.filePath, data, 0644)
	if err != nil {
		return errors.New("写入历史记录文件失败: " + err.Error())
	}

	return nil
}

// GetHistory 获取最近的命令历史记录
func (h *FileCommandHistory) GetHistory(limit int) ([]HistoryRecord, error) {
	if limit <= 0 || limit > len(h.records) {
		limit = len(h.records)
	}

	start := len(h.records) - limit
	if start < 0 {
		start = 0
	}

	return h.records[start:], nil
} 