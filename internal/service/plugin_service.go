package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"novel-agent-os-backend/internal/model"
	"novel-agent-os-backend/internal/repository"
	"novel-agent-os-backend/pkg/logger"
)

type PluginService interface {
	CreatePlugin(plugin *model.Plugin) error
	GetPlugin(id uint) (*model.Plugin, error)
	GetPluginByName(name string) (*model.Plugin, error)
	UpdatePlugin(plugin *model.Plugin) error
	DeletePlugin(id uint) error
	ListPlugins(page, pageSize int) ([]*model.Plugin, int64, error)
	ListEnabledPlugins() ([]*model.Plugin, error)

	EnablePlugin(id uint) error
	DisablePlugin(id uint) error

	UpdatePluginHealth(id uint, healthy bool, latencyMs int) error
	PingPlugin(id uint) error

	AddCapability(capability *model.PluginCapability) error
	GetCapabilities(pluginID uint) ([]*model.PluginCapability, error)
	RemoveCapability(id uint) error

	InvokePlugin(ctx context.Context, id uint, method string, payload map[string]interface{}, authorizationHeader string) (*PluginInvokeResult, error)
}

// PluginInvokeResult 插件调用结果
type PluginInvokeResult struct {
	Success  bool                   `json:"success"`
	Data     map[string]interface{} `json:"data"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

type pluginService struct {
	pluginRepo repository.PluginRepository
}

type pluginInvokeRequest struct {
	Method  string                 `json:"method"`
	Payload map[string]interface{} `json:"payload"`
}

func NewPluginService(pluginRepo repository.PluginRepository) PluginService {
	return &pluginService{
		pluginRepo: pluginRepo,
	}
}

func (s *pluginService) CreatePlugin(plugin *model.Plugin) error {
	return s.pluginRepo.Create(plugin)
}

func (s *pluginService) GetPlugin(id uint) (*model.Plugin, error) {
	return s.pluginRepo.GetByID(id)
}

func (s *pluginService) GetPluginByName(name string) (*model.Plugin, error) {
	return s.pluginRepo.GetByName(name)
}

func (s *pluginService) UpdatePlugin(plugin *model.Plugin) error {
	return s.pluginRepo.Update(plugin)
}

func (s *pluginService) DeletePlugin(id uint) error {
	return s.pluginRepo.Delete(id)
}

func (s *pluginService) ListPlugins(page, pageSize int) ([]*model.Plugin, int64, error) {
	return s.pluginRepo.List(page, pageSize)
}

func (s *pluginService) ListEnabledPlugins() ([]*model.Plugin, error) {
	return s.pluginRepo.ListEnabled()
}

func (s *pluginService) EnablePlugin(id uint) error {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return err
	}
	plugin.IsEnabled = true
	plugin.Status = "enabled"
	return s.pluginRepo.Update(plugin)
}

func (s *pluginService) DisablePlugin(id uint) error {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return err
	}
	plugin.IsEnabled = false
	plugin.Status = "disabled"
	return s.pluginRepo.Update(plugin)
}

func (s *pluginService) UpdatePluginHealth(id uint, healthy bool, latencyMs int) error {
	return s.pluginRepo.UpdateHealth(id, healthy, latencyMs)
}

func (s *pluginService) PingPlugin(id uint) error {
	return s.pluginRepo.UpdateLastPing(id)
}

func (s *pluginService) AddCapability(capability *model.PluginCapability) error {
	return s.pluginRepo.CreateCapability(capability)
}

func (s *pluginService) GetCapabilities(pluginID uint) ([]*model.PluginCapability, error) {
	return s.pluginRepo.GetCapabilitiesByPluginID(pluginID)
}

func (s *pluginService) RemoveCapability(id uint) error {
	return s.pluginRepo.DeleteCapability(id)
}

func (s *pluginService) InvokePlugin(ctx context.Context, id uint, method string, payload map[string]interface{}, authorizationHeader string) (*PluginInvokeResult, error) {
	plugin, err := s.pluginRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if !plugin.IsEnabled {
		return nil, fmt.Errorf("插件已禁用")
	}

	endpoint := strings.TrimSpace(plugin.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("插件未配置 endpoint")
	}

	targetURL, err := buildPluginInvokeURL(endpoint)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(pluginInvokeRequest{Method: method, Payload: payload})
	if err != nil {
		return nil, fmt.Errorf("序列化插件请求失败: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建插件请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(authorizationHeader) != "" {
		httpReq.Header.Set("Authorization", authorizationHeader)
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	latencyMs := int(time.Since(start).Milliseconds())

	// 无论成功/失败，都尝试更新健康状态（不影响主流程返回）
	defer func() {
		healthy := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
		if uErr := s.pluginRepo.UpdateHealth(id, healthy, latencyMs); uErr != nil {
			logger.Warn("更新插件健康状态失败", logger.Err(uErr), logger.Uint("plugin_id", id))
		}
	}()

	if err != nil {
		logger.Error("插件调用失败", logger.Err(err), logger.Uint("plugin_id", id), logger.String("url", targetURL))
		return nil, fmt.Errorf("插件调用失败: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("读取插件响应失败: %w", readErr)
	}

	// 非 2xx 认为调用失败，尽量把响应内容带出来便于排查
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("插件返回非成功状态: %s", msg)
	}

	data := map[string]interface{}{}
	if len(raw) > 0 {
		if jErr := json.Unmarshal(raw, &data); jErr != nil {
			// 插件返回不一定是 JSON 对象，兜底保留原始文本
			data = map[string]interface{}{
				"raw": string(raw),
			}
		}
	}

	return &PluginInvokeResult{
		Success:  true,
		Data:     data,
		Metadata: map[string]interface{}{"latency_ms": latencyMs, "url": targetURL},
	}, nil
}

func buildPluginInvokeURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("解析插件 endpoint 失败: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("插件 endpoint 必须是完整 URL（例如 http://127.0.0.1:9000）")
	}

	// 约定：如果 endpoint 只给到 host（path 为空或 /），默认调用 /invoke
	if u.Path == "" || u.Path == "/" {
		u.Path = "/invoke"
	}
	return u.String(), nil
}
