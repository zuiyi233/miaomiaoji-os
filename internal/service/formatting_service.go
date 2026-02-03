package service

import (
	"regexp"
	"strings"

	"novel-agent-os-backend/internal/config"
)

type FormattingService interface {
	FormatText(text string, style string) (string, error)
	GetAvailableStyles() []string
}

type formattingService struct {
	cfg config.Config
}

func NewFormattingService(cfg config.Config) FormattingService {
	return &formattingService{
		cfg: cfg,
	}
}

func (s *formattingService) FormatText(text string, style string) (string, error) {
	switch style {
	case "tomato":
		return s.formatTomato(text)
	case "standard":
		return s.formatStandard(text)
	default:
		return text, nil
	}
}

func (s *formattingService) formatTomato(text string) (string, error) {
	lines := strings.Split(text, "\n")
	var paragraphs []string
	var currentParagraph strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if currentParagraph.Len() > 0 {
				paragraphs = append(paragraphs, currentParagraph.String())
				currentParagraph.Reset()
			}
			continue
		}

		if currentParagraph.Len() > 0 {
			currentParagraph.WriteString(" ")
		}
		currentParagraph.WriteString(trimmed)
	}

	if currentParagraph.Len() > 0 {
		paragraphs = append(paragraphs, currentParagraph.String())
	}

	var result strings.Builder
	for i, para := range paragraphs {
		if i > 0 {
			result.WriteString("\n\n")
		}

		trimmed := strings.TrimSpace(para)
		if s.isHeading(trimmed) {
			result.WriteString(trimmed)
		} else {
			result.WriteString("　　")
			result.WriteString(trimmed)
		}
	}

	return result.String(), nil
}

func (s *formattingService) formatStandard(text string) (string, error) {
	result := s.convertPunctuationToFullwidth(text)
	result = s.fixSpacing(result)
	return result, nil
}

func (s *formattingService) isHeading(text string) bool {
	if len(text) < 2 || len(text) > 20 {
		return false
	}

	headingPattern := regexp.MustCompile(`^(第[一二三四五六七八九十百千]+章|第\d+章|Chapter\s+\d+|\d+\.\s*|[一二三四五六七八九十]、)`)
	if headingPattern.MatchString(text) {
		return true
	}

	if strings.HasSuffix(text, ":") || strings.HasSuffix(text, "：") {
		return true
	}

	if len(text) < 10 && !strings.Contains(text, "。") && !strings.Contains(text, "，") {
		return true
	}

	return false
}

func (s *formattingService) convertPunctuationToFullwidth(text string) string {
	replacements := map[rune]rune{
		',':  '，',
		'.':  '。',
		':':  '：',
		';':  '；',
		'?':  '？',
		'!':  '！',
		'(':  '（',
		')':  '）',
		'[':  '［',
		']':  '］',
		'{':  '｛',
		'}':  '｝',
		'"':  '"',
		'\'': '\'',
		'<':  '《',
		'>':  '》',
	}

	var result strings.Builder
	for _, r := range text {
		if replaced, ok := replacements[r]; ok {
			result.WriteRune(replaced)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (s *formattingService) fixSpacing(text string) string {
	text = strings.ReplaceAll(text, "  ", " ")
	text = strings.ReplaceAll(text, "　", " ")
	text = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(text, "\n\n")

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n\n")
}

func (s *formattingService) GetAvailableStyles() []string {
	return []string{"tomato", "standard"}
}
