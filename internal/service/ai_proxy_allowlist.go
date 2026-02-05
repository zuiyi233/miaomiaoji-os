package service

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// ValidateAIProxyTarget 校验 provider+path 是否允许被代理到上游。
//
// 约束目标：
// 1) 禁止绝对 URL（避免绕过 BaseURL）。
// 2) BaseURL 仅允许 https；http 仅允许 localhost/127.0.0.1/::1（便于本地调试）。
// 3) path 必须命中白名单（最小集合，避免任意上游路径访问）。
// ValidateAIProxyTarget 导出以供 handler/service 复用。
func ValidateAIProxyTarget(provider, baseURL, path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("invalid path")
	}
	if strings.Contains(path, "..") || strings.Contains(path, "\\") {
		return errors.New("invalid path")
	}

	// 禁止携带 scheme/host 的绝对 URL
	parsedPath, err := url.Parse(path)
	if err != nil {
		return errors.New("invalid path")
	}
	if parsedPath.IsAbs() || parsedPath.Host != "" {
		return errors.New("invalid path")
	}
	// gemini 流式接口需要 query（alt=sse），其他场景不允许携带 query/fragment
	cleanProvider := strings.ToLower(strings.TrimSpace(provider))
	if parsedPath.Fragment != "" {
		return errors.New("invalid path")
	}
	if parsedPath.RawQuery != "" {
		if cleanProvider != "gemini" {
			return errors.New("invalid path")
		}
		if parsedPath.RawQuery != "alt=sse" {
			return errors.New("invalid path")
		}
	}

	// BaseURL scheme 校验
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base.Scheme == "" {
		return errors.New("invalid base url")
	}
	if base.Scheme != "https" {
		if base.Scheme != "http" {
			return errors.New("invalid base url")
		}
		host := base.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			// 生产环境禁用明文 http
			return errors.New("invalid base url")
		}
	}

	normalized := "/" + strings.TrimLeft(parsedPath.Path, "/")
	if normalized == "/" {
		return errors.New("invalid path")
	}

	// provider 白名单：
	// - gemini：仅放行 generateContent/streamGenerateContent
	// - 其他：按 OpenAI 兼容接口放行（workflow 依赖 chat/completions 注入 tools）
	if cleanProvider == "gemini" {
		if !strings.HasPrefix(normalized, "/v1beta/models/") {
			return errors.New("path not allowed")
		}
		if !strings.Contains(normalized, ":generateContent") && !strings.Contains(normalized, ":streamGenerateContent") {
			return errors.New("path not allowed")
		}
		return nil
	}

	allowedExact := map[string]struct{}{
		"/v1/chat/completions": {},
		"/chat/completions":    {},
		"/v1/completions":      {},
		"/completions":         {},
		"/v1/embeddings":       {},
		"/embeddings":          {},
		"/v1/responses":        {},
		"/responses":           {},
	}
	if _, ok := allowedExact[normalized]; ok {
		return nil
	}

	// 兼容：上游可能把版本前缀放在 BaseURL（例如 baseURL=.../v1, path=/chat/completions）
	if strings.HasSuffix(strings.TrimRight(base.Path, "/"), "/v1") {
		if strings.HasPrefix(normalized, "/chat/") || normalized == "/completions" || normalized == "/embeddings" || normalized == "/responses" {
			return nil
		}
	}

	// 额外保护：阻断内网探测（如果 baseURL 指向内网域名但 https 仍可能被误配）
	// 这里以保守策略拒绝明显的私网 IP。
	if ip := net.ParseIP(base.Hostname()); ip != nil {
		if ip.IsPrivate() {
			return errors.New("invalid base url")
		}
	}

	return errors.New("path not allowed")
}
