package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	BaseURL = "http://localhost:8080"
	APIBase = "http://192.168.32.15:39999/v1"  // Must include /v1
	APIKey  = "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B"
	Model   = "glm-4.7"
)

type TestResult struct {
	Name      string    `json:"name"`
	Success   bool      `json:"success"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

var testResults []TestResult

func logTest(name string, success bool, details string) {
	result := TestResult{
		Name:      name,
		Success:   success,
		Details:   details,
		Timestamp: time.Now(),
	}
	testResults = append(testResults, result)
	status := "[PASS]"
	if !success {
		status = "[FAIL]"
	}
	fmt.Printf("  %s %s: %s\n", status, name, details)
}

func makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
	url := BaseURL + endpoint
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	return client.Do(req)
}

func parseResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Test 1: Health Check
func testHealthCheck() {
	fmt.Println("\n[Test] 服务健康检查")
	resp, err := makeRequest("GET", "/healthz", nil, "")
	if err != nil {
		logTest("健康检查", false, err.Error())
		return
	}

	result, err := parseResponse(resp)
	if err != nil {
		logTest("健康检查", false, err.Error())
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if ok && data["status"] == "ok" {
		logTest("健康检查", true, "服务运行正常")
	} else {
		logTest("健康检查", false, "服务状态异常")
	}
}

// Test 2: User Auth
func testUserAuth() string {
	fmt.Println("\n[Test] 用户登录 (使用默认admin账户)")

	// Use default admin account
	// Default password is "admin"
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin",
	}

	resp, err := makeRequest("POST", "/api/v1/auth/login", loginBody, "")
	if err != nil {
		logTest("用户登录", false, err.Error())
		return ""
	}

	result, err := parseResponse(resp)
	if err != nil {
		logTest("用户登录", false, err.Error())
		return ""
	}

	code, _ := result["code"].(float64)
	if code != 0 {
		msg, _ := result["message"].(string)
		logTest("用户登录", false, "登录失败: "+msg)
		return ""
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		logTest("用户登录", false, "响应格式错误: data字段不存在")
		return ""
	}

	token, ok := data["token"].(string)
	if !ok || token == "" {
		logTest("用户登录", false, "未获取到token")
		return ""
	}

	logTest("用户登录", true, "Token获取成功 (admin)")
	return token
}

// Test 3: AI Config
func testAIConfig(token string) {
	fmt.Println("\n[Test] AI配置管理")

	// Update provider config - use the correct API key
	providerBody := map[string]string{
		"provider": "zhipu",
		"base_url": "http://192.168.32.15:39999/v1",
		"api_key":  "sk-jUIhuUtm36APmZhqvrkmjZOU4bRwDApQMfUQehK8Z2vht60B",
	}

	resp, err := makeRequest("PUT", "/api/v1/ai/providers", providerBody, token)
	if err != nil {
		logTest("更新供应商配置", false, err.Error())
		return
	}

	_, err = parseResponse(resp)
	if err != nil {
		logTest("更新供应商配置", false, err.Error())
		return
	}
	logTest("更新供应商配置", true, "Provider: zhipu")

	// Get provider config
	resp, err = makeRequest("GET", "/api/v1/ai/providers?provider=zhipu", nil, token)
	if err != nil {
		logTest("获取供应商配置", false, err.Error())
		return
	}

	result, err := parseResponse(resp)
	if err != nil {
		logTest("获取供应商配置", false, err.Error())
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if ok && data["provider"] == "zhipu" {
		logTest("获取供应商配置", true, "BaseURL: "+data["base_url"].(string))
	} else {
		logTest("获取供应商配置", false, "配置获取失败")
	}

	// Test provider connection
	testBody := map[string]string{"provider": "zhipu"}
	resp, err = makeRequest("POST", "/api/v1/ai/providers/test", testBody, token)
	if err != nil {
		logTest("测试供应商连接", false, err.Error())
		return
	}

	result, err = parseResponse(resp)
	if err != nil {
		logTest("测试供应商连接", false, err.Error())
		return
	}

	code, _ := result["code"].(float64)
	if code == 0 {
		logTest("测试供应商连接", true, "连接成功")
	} else {
		msg, _ := result["message"].(string)
		logTest("测试供应商连接", false, "连接失败: "+msg)
	}
}

// Test 4: Model List
func testModelList(token string) {
	fmt.Println("\n[Test] AI模型列表")

	resp, err := makeRequest("GET", "/api/v1/ai/models?provider=zhipu", nil, token)
	if err != nil {
		logTest("获取模型列表", false, err.Error())
		return
	}

	result, err := parseResponse(resp)
	if err != nil {
		logTest("获取模型列表", false, err.Error())
		return
	}

	data, ok := result["data"].([]interface{})
	if ok && len(data) > 0 {
		modelNames := []string{}
		for i, m := range data {
			if i >= 5 {
				break
			}
			model := m.(map[string]interface{})
			modelNames = append(modelNames, model["id"].(string))
		}
		logTest("获取模型列表", true, fmt.Sprintf("找到 %d 个模型: %s", len(data), strings.Join(modelNames, ", ")))
	} else {
		logTest("获取模型列表", false, "模型列表为空或获取失败")
	}
}

// Test 5: AI Chat
func testAIChat(token string) {
	fmt.Println("\n[Test] AI普通对话")
	fmt.Println("  发送请求到 /v1/chat/completions...")

	chatBody := map[string]interface{}{
		"provider": "zhipu",
		"path":     "/v1/chat/completions",
		"body": map[string]interface{}{
			"model": Model,
			"messages": []map[string]string{
				{"role": "system", "content": "你是一个 helpful assistant."},
				{"role": "user", "content": "你好，请简单介绍一下自己"},
			},
			"temperature": 0.7,
			"max_tokens":  200,
		},
	}

	resp, err := makeRequest("POST", "/api/v1/ai/proxy", chatBody, token)
	if err != nil {
		logTest("AI普通对话", false, err.Error())
		return
	}

	// Parse as raw JSON since it's proxied from AI provider
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logTest("AI普通对话", false, err.Error())
		return
	}

	var aiResp map[string]interface{}
	if err := json.Unmarshal(body, &aiResp); err != nil {
		logTest("AI普通对话", false, "解析响应失败: "+err.Error())
		return
	}

	choices, ok := aiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		logTest("AI普通对话", false, "响应中没有choices")
		return
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)
	model := aiResp["model"].(string)

	preview := content
	if len(preview) > 80 {
		preview = preview[:80] + "..."
	}
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  响应: %s\n", preview)
	logTest("AI普通对话", true, fmt.Sprintf("Model: %s, 响应长度: %d", model, len(content)))
}

// Test 6: AI Stream
func testAIStream(token string) {
	fmt.Println("\n[Test] AI流式对话")
	fmt.Println("  发送流式请求...")

	streamBody := map[string]interface{}{
		"provider": "zhipu",
		"path":     "/v1/chat/completions",
		"body": map[string]interface{}{
			"model": Model,
			"messages": []map[string]string{
				{"role": "user", "content": "用一句话描述春天的美丽"},
			},
			"stream":     true,
			"max_tokens": 100,
		},
	}

	url := BaseURL + "/api/v1/ai/proxy/stream"
	jsonBody, _ := json.Marshal(streamBody)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		logTest("AI流式对话", false, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logTest("AI流式对话", false, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		logTest("AI流式对话", true, "流式响应接收成功")
	} else {
		logTest("AI流式对话", false, fmt.Sprintf("状态码: %d", resp.StatusCode))
	}
}

// Test 7: Workflow
func testWorkflow(token string) {
	fmt.Println("\n[Test] 工作流功能")

	// Create project
	projectBody := map[string]string{
		"title":       fmt.Sprintf("测试项目_%d", time.Now().Unix()),
		"description": "AI测试项目",
	}

	resp, err := makeRequest("POST", "/api/v1/projects/", projectBody, token)
	if err != nil {
		logTest("创建项目", false, err.Error())
		return
	}

	result, err := parseResponse(resp)
	if err != nil {
		logTest("创建项目", false, err.Error())
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		logTest("创建项目", false, "响应格式错误")
		return
	}

	projectID := uint(data["id"].(float64))
	logTest("创建项目", true, fmt.Sprintf("ProjectID: %d", projectID))

	// World building workflow
	fmt.Println("  执行世界构建工作流...")
	worldBody := map[string]interface{}{
		"project_id":     projectID,
		"session_title":  "世界构建测试",
		"prompt":         "创建一个简单的奇幻世界设定，包含世界名称和基本描述",
		"provider":       "zhipu",
		"model":          Model,
	}

	resp, err = makeRequest("POST", "/api/v1/workflows/world", worldBody, token)
	if err != nil {
		logTest("世界构建工作流", false, err.Error())
		return
	}

	result, err = parseResponse(resp)
	if err != nil {
		logTest("世界构建工作流", false, err.Error())
		return
	}

	code, _ := result["code"].(float64)
	if code == 0 {
		respData := result["data"].(map[string]interface{})
		session := respData["session"].(map[string]interface{})
		logTest("世界构建工作流", true, fmt.Sprintf("SessionID: %d", uint(session["id"].(float64))))
	} else {
		msg, _ := result["message"].(string)
		logTest("世界构建工作流", false, msg)
	}
}

// Test 8: Error Handling
func testErrorHandling(token string) {
	fmt.Println("\n[Test] 错误处理测试")

	// Test invalid provider
	resp, err := makeRequest("GET", "/api/v1/ai/models?provider=nonexistent", nil, token)
	if err != nil {
		logTest("无效供应商处理", true, "正确返回错误")
	} else {
		result, _ := parseResponse(resp)
		code, _ := result["code"].(float64)
		if code != 0 {
			logTest("无效供应商处理", true, "正确返回错误")
		} else {
			logTest("无效供应商处理", false, "应该返回错误")
		}
	}

	// Test unauthorized access
	resp, err = makeRequest("GET", "/api/v1/ai/models?provider=zhipu", nil, "")
	if err != nil {
		logTest("未授权访问处理", true, "正确拒绝未授权请求")
	} else {
		logTest("未授权访问处理", false, "应该拒绝未授权请求")
	}
}

func printReport() {
	fmt.Println("\n========================================")
	fmt.Println("  测试报告汇总")
	fmt.Println("========================================")

	total := len(testResults)
	passed := 0
	for _, r := range testResults {
		if r.Success {
			passed++
		}
	}
	failed := total - passed

	fmt.Printf("总测试数: %d\n", total)
	fmt.Printf("通过: %d\n", passed)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("成功率: %.2f%%\n", float64(passed)/float64(total)*100)

	fmt.Println("\n详细结果:")
	for _, r := range testResults {
		status := "[PASS]"
		if !r.Success {
			status = "[FAIL]"
		}
		fmt.Printf("  %s %s - %s\n", status, r.Name, r.Details)
	}

	// Save report
	reportData, _ := json.MarshalIndent(testResults, "", "  ")
	reportPath := fmt.Sprintf("test/ai_test/test_report_%s.json", time.Now().Format("20060102_150405"))
	os.WriteFile(reportPath, reportData, 0644)
	fmt.Printf("\n测试报告已保存到: %s\n", reportPath)
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  AI功能全量测试开始")
	fmt.Println("========================================")
	fmt.Printf("API: %s\n", APIBase)
	fmt.Printf("Model: %s\n", Model)
	fmt.Println("========================================")

	// 1. Health Check
	testHealthCheck()

	// 2. User Auth
	token := testUserAuth()
	if token == "" {
		fmt.Println("\n无法获取Token，终止测试")
		printReport()
		os.Exit(1)
	}

	// 3. AI Config
	testAIConfig(token)

	// 4. Model List
	testModelList(token)

	// 5. AI Chat
	testAIChat(token)

	// 6. AI Stream
	testAIStream(token)

	// 7. Workflow
	testWorkflow(token)

	// 8. Error Handling
	testErrorHandling(token)

	// Print report
	printReport()
}
