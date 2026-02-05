// Package service 质量门禁服务
package service

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"novel-agent-os-backend/internal/config"
	"novel-agent-os-backend/pkg/logger"
)

// QualityCheckResult 质量检查结果
type QualityCheckResult struct {
	Passed  bool                   `json:"passed"`
	Score   int                    `json:"score"`
	Issues  []QualityIssue         `json:"issues"`
	Details map[string]interface{} `json:"details"`
}

// QualityIssue 质量问题项
type QualityIssue struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Position int    `json:"position"`
}

// QualityGateService 质量门禁服务接口
type QualityGateService interface {
	CheckQuality(content string) (*QualityCheckResult, error)
	GetThresholds() map[string]int
}

type qualityGateService struct {
	cfg config.Config
}

// NewQualityGateService 创建质量门禁服务
func NewQualityGateService(cfg config.Config) QualityGateService {
	return &qualityGateService{
		cfg: cfg,
	}
}

// CheckQuality 检查内容质量
func (s *qualityGateService) CheckQuality(content string) (*QualityCheckResult, error) {
	logger.Debug("开始质量检查", logger.Int("content_length", len(content)))

	result := &QualityCheckResult{
		Passed:  true,
		Score:   100,
		Issues:  []QualityIssue{},
		Details: make(map[string]interface{}),
	}

	// 基础统计
	totalChars := len(content)
	nonSpaceChars := s.countNonSpaceChars(content)
	paragraphs := s.splitParagraphs(content)
	avgParaLength := 0
	if len(paragraphs) > 0 {
		avgParaLength = nonSpaceChars / len(paragraphs)
	}

	result.Details["total_chars"] = totalChars
	result.Details["non_space_chars"] = nonSpaceChars
	result.Details["paragraph_count"] = len(paragraphs)
	result.Details["avg_paragraph_length"] = avgParaLength

	// 检查1: 最小长度
	minLength := s.getMinLength()
	if nonSpaceChars < minLength {
		result.Score -= 20
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "length",
			Severity: "high",
			Message:  "内容过短，建议至少 " + strconv.Itoa(minLength) + " 个有效字符",
		})
		result.Passed = false
	}

	// 检查2: 最大长度
	maxLength := s.getMaxLength()
	if nonSpaceChars > maxLength {
		result.Score -= 10
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "length",
			Severity: "medium",
			Message:  "内容过长，建议控制在 " + strconv.Itoa(maxLength) + " 个字符以内",
		})
	}

	// 检查3: 段落长度检查
	s.checkParagraphLengths(paragraphs, result)

	// 检查4: 钩子检测（开头是否有吸引力）
	s.checkHook(content, result)

	// 检查5: 标点符号检查
	s.checkPunctuation(content, result)

	// 检查6: 重复内容检查
	s.checkRepetition(content, result)

	// 检查7: 敏感词检查
	s.checkSensitiveWords(content, result)

	// 确保分数不低于0
	if result.Score < 0 {
		result.Score = 0
	}

	result.Details["score_breakdown"] = map[string]int{
		"final_score":  result.Score,
		"issues_count": len(result.Issues),
	}

	logger.Info("质量检查完成",
		logger.Int("score", result.Score),
		logger.Bool("passed", result.Passed),
		logger.Int("issues", len(result.Issues)))

	return result, nil
}

// GetThresholds 获取检查阈值
func (s *qualityGateService) GetThresholds() map[string]int {
	return map[string]int{
		"min_length":       s.getMinLength(),
		"max_length":       s.getMaxLength(),
		"min_paragraph":    s.getMinParagraphLength(),
		"max_paragraph":    s.getMaxParagraphLength(),
		"min_paragraphs":   1,
		"max_repetition":   30,
		"hook_check_chars": 200,
	}
}

// 段落长度检查
func (s *qualityGateService) checkParagraphLengths(paragraphs []string, result *QualityCheckResult) {
	minPara := s.getMinParagraphLength()
	maxPara := s.getMaxParagraphLength()

	for i, para := range paragraphs {
		paraLen := s.countNonSpaceChars(para)

		if paraLen < minPara && paraLen > 0 {
			result.Score -= 5
			result.Issues = append(result.Issues, QualityIssue{
				Type:     "paragraph",
				Severity: "low",
				Message:  "段落 " + strconv.Itoa(i+1) + " 过短，建议充实内容",
				Position: i + 1,
			})
		}

		if paraLen > maxPara {
			result.Score -= 5
			result.Issues = append(result.Issues, QualityIssue{
				Type:     "paragraph",
				Severity: "low",
				Message:  "段落 " + strconv.Itoa(i+1) + " 过长，建议适当分段",
				Position: i + 1,
			})
		}
	}
}

// 钩子检测（检查开头是否有吸引力）
func (s *qualityGateService) checkHook(content string, result *QualityCheckResult) {
	hookChars := 200
	if len(content) < hookChars {
		hookChars = len(content)
	}

	hookSection := content[:hookChars]

	// 检查开头是否有对话或动作描写
	hasDialogue := strings.Contains(hookSection, "「") || strings.Contains(hookSection, "」") ||
		strings.Contains(hookSection, "'") || strings.Contains(hookSection, "\"")
	hasAction := strings.Contains(hookSection, "！") || strings.Contains(hookSection, "？")

	if !hasDialogue && !hasAction {
		result.Score -= 10
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "hook",
			Severity: "medium",
			Message:  "开头缺乏吸引力，建议加入对话或动作描写吸引读者",
		})
	}

	// 检查开头是否过于平淡（只有描述性文字）
	if strings.HasPrefix(strings.TrimSpace(content), "那是") ||
		strings.HasPrefix(strings.TrimSpace(content), "这是") ||
		strings.HasPrefix(strings.TrimSpace(content), "有") {
		result.Score -= 5
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "hook",
			Severity: "low",
			Message:  "开头较为平淡，建议用更有冲击力的场景开头",
		})
	}
}

// 标点符号检查
func (s *qualityGateService) checkPunctuation(content string, result *QualityCheckResult) {
	// 检查标点使用是否规范
	if strings.Contains(content, ",,") || strings.Contains(content, "，，") {
		result.Score -= 5
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "punctuation",
			Severity: "medium",
			Message:  "存在连续标点符号，请检查标点使用",
		})
	}

	// 检查中英文标点混用
	halfWidthPunct := regexp.MustCompile(`[,.?!;:'"()[\]{}<>]`)
	if halfWidthPunct.MatchString(content) {
		result.Score -= 5
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "punctuation",
			Severity: "low",
			Message:  "建议使用全角标点符号",
		})
	}

	// 检查句号缺失
	if !strings.Contains(content, "。") && !strings.Contains(content, "！") && !strings.Contains(content, "？") {
		result.Score -= 10
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "punctuation",
			Severity: "high",
			Message:  "内容缺少标点符号，请添加适当的句号、问号或感叹号",
		})
	}

	// 检查连续感叹号或问号
	if strings.Contains(content, "！！") || strings.Contains(content, "？？") {
		result.Score -= 3
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "punctuation",
			Severity: "low",
			Message:  "避免连续使用感叹号或问号",
		})
	}
}

// 重复内容检查
func (s *qualityGateService) checkRepetition(content string, result *QualityCheckResult) {
	// 简单重复检测 - 检查是否有大量重复的词语
	words := s.extractWords(content)
	if len(words) == 0 {
		return
	}

	wordCount := make(map[string]int)
	for _, word := range words {
		if len(word) >= 2 {
			wordCount[word]++
		}
	}

	// 找出高频重复词
	repeatedWords := []string{}
	threshold := len(words) / 20 // 如果某个词出现超过5%的文本量
	if threshold < 5 {
		threshold = 5
	}

	for word, count := range wordCount {
		if count > threshold {
			repeatedWords = append(repeatedWords, word)
		}
	}

	if len(repeatedWords) > 0 {
		result.Score -= 8
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "repetition",
			Severity: "medium",
			Message:  "文本中存在高频重复词汇，建议使用同义词替换",
		})
	}

	result.Details["unique_word_count"] = len(wordCount)
}

// 敏感词检查
func (s *qualityGateService) checkSensitiveWords(content string, result *QualityCheckResult) {
	// 基础敏感词列表（示例）
	sensitiveWords := []string{
		"色情", "暴力", "毒品", "赌博", "反动", "政治",
	}

	foundWords := []string{}
	lowerContent := strings.ToLower(content)

	for _, word := range sensitiveWords {
		if strings.Contains(lowerContent, word) {
			foundWords = append(foundWords, word)
		}
	}

	if len(foundWords) > 0 {
		result.Score -= 50
		result.Passed = false
		result.Issues = append(result.Issues, QualityIssue{
			Type:     "sensitive",
			Severity: "critical",
			Message:  "内容包含敏感词汇，请修改后重新提交",
		})
	}
}

// 辅助方法
func (s *qualityGateService) countNonSpaceChars(text string) int {
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

func (s *qualityGateService) splitParagraphs(text string) []string {
	// 根据换行符分割段落
	paragraphs := regexp.MustCompile(`\n\s*\n`).Split(text, -1)
	result := []string{}
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			result = append(result, p)
		}
	}
	return result
}

func (s *qualityGateService) extractWords(text string) []string {
	// 简单的中文分词（基于字符和标点）
	var words []string
	currentWord := ""

	for _, r := range text {
		if unicode.Is(unicode.Han, r) || unicode.IsLetter(r) {
			currentWord += string(r)
		} else {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// 配置获取方法
func (s *qualityGateService) getMinLength() int {
	return 100
}

func (s *qualityGateService) getMaxLength() int {
	return 50000
}

func (s *qualityGateService) getMinParagraphLength() int {
	return 50
}

func (s *qualityGateService) getMaxParagraphLength() int {
	return 800
}
