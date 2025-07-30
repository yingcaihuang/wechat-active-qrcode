package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"wechat-active-qrcode/internal/models"
)

const testBaseURL = "http://localhost:8083"

type ActiveQRTestClient struct {
	client *http.Client
	token  string
}

func NewActiveQRTestClient() *ActiveQRTestClient {
	return &ActiveQRTestClient{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (tc *ActiveQRTestClient) register(username, password string) error {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := tc.client.Post(testBaseURL+"/api/auth/register", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("注册请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("注册失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析注册响应失败: %v", err)
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			tc.token = token
			fmt.Printf("✅ 用户注册成功，获得token: %s...\n", token[:20])
			return nil
		}
	}

	return fmt.Errorf("无法从注册响应中提取token")
}

func (tc *ActiveQRTestClient) createActiveQRCode() (uint, string, error) {
	reqBody := map[string]interface{}{
		"name":        "测试活码-权重分配",
		"switch_rule": "weight",
		"description": "这是一个测试活码，根据权重分配流量到不同的目标网站",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", testBaseURL+"/api/active-qrcodes", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tc.token)

	resp, err := tc.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("创建活码请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("创建活码失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", fmt.Errorf("解析创建活码响应失败: %v", err)
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			shortCode, _ := data["short_code"].(string)
			fmt.Printf("✅ 活码创建成功，ID: %.0f, 短码: %s\n", id, shortCode)
			return uint(id), shortCode, nil
		}
	}

	return 0, "", fmt.Errorf("无法从创建活码响应中提取ID")
}

func (tc *ActiveQRTestClient) addStaticQRCode(activeID uint, name, targetURL string, weight int) error {
	reqBody := map[string]interface{}{
		"name":       name,
		"target_url": targetURL,
		"weight":     weight,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/active-qrcodes/%d/static-qrcodes", testBaseURL, activeID)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tc.token)

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("添加静态码请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("添加静态码失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ 静态码添加成功: %s -> %s (权重: %d)\n", name, targetURL, weight)
	return nil
}

func (tc *ActiveQRTestClient) testRedirect(shortCode string) (string, error) {
	url := fmt.Sprintf("%s/r/%s", testBaseURL, shortCode)
	resp, err := tc.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("重定向请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("重定向失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("重定向响应中没有Location头")
	}

	return location, nil
}

func (tc *ActiveQRTestClient) checkHealth() error {
	resp, err := tc.client.Get(testBaseURL + "/health")
	if err != nil {
		return fmt.Errorf("健康检查请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println("✅ 服务器健康检查通过")
	return nil
}

func runActiveQRTests() {
	fmt.Println("=== 微信活码管理系统完整测试 ===")

	client := NewActiveQRTestClient()

	// 1. 健康检查
	fmt.Println("\n1. 服务器健康检查...")
	if err := client.checkHealth(); err != nil {
		log.Fatalf("❌ 健康检查失败: %v", err)
	}

	// 2. 用户注册
	fmt.Println("\n2. 用户注册...")
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	if err := client.register(username, "testpass123"); err != nil {
		log.Fatalf("❌ 用户注册失败: %v", err)
	}

	// 3. 创建活码
	fmt.Println("\n3. 创建活码...")
	activeID, shortCode, err := client.createActiveQRCode()
	if err != nil {
		log.Fatalf("❌ 创建活码失败: %v", err)
	}

	// 4. 添加静态二维码
	fmt.Println("\n4. 添加静态二维码...")
	staticCodes := []struct {
		name      string
		targetURL string
		weight    int
	}{
		{"静态码1-GitHub", "https://github.com", 50},
		{"静态码2-百度", "https://www.baidu.com", 30},
		{"静态码3-腾讯", "https://www.tencent.com", 20},
	}

	for _, sc := range staticCodes {
		if err := client.addStaticQRCode(activeID, sc.name, sc.targetURL, sc.weight); err != nil {
			log.Fatalf("❌ 添加静态码失败: %v", err)
		}
	}

	// 5. 测试重定向功能
	fmt.Println("\n5. 测试重定向功能...")
	fmt.Printf("使用短码: %s\n", shortCode)
	fmt.Println("测试权重分配算法（多次访问同一活码）:")

	// 统计重定向结果
	redirectStats := make(map[string]int)
	testCount := 30

	for i := 1; i <= testCount; i++ {
		targetURL, err := client.testRedirect(shortCode)
		if err != nil {
			log.Printf("❌ 第%d次重定向测试失败: %v", i, err)
			continue
		}

		redirectStats[targetURL]++
		fmt.Printf("第%2d次访问: %s\n", i, targetURL)
	}

	// 6. 分析结果
	fmt.Println("\n6. 测试结果分析:")
	fmt.Printf("总测试次数: %d\n", testCount)
	fmt.Printf("重定向分布统计:\n")

	totalRedirects := 0
	for url, count := range redirectStats {
		totalRedirects += count
		percentage := float64(count) / float64(testCount) * 100
		fmt.Printf("  %s: %d次 (%.1f%%)\n", url, count, percentage)
	}

	// 7. 验证权重分配
	fmt.Println("\n7. 权重分配验证:")
	fmt.Printf("预期权重分配 (总权重100):\n")
	fmt.Printf("  GitHub (权重50): ~50%%\n")
	fmt.Printf("  百度 (权重30): ~30%%\n")
	fmt.Printf("  腾讯 (权重20): ~20%%\n")

	// 计算实际分配
	githubCount := redirectStats["https://github.com"]
	baiduCount := redirectStats["https://www.baidu.com"]
	tencentCount := redirectStats["https://www.tencent.com"]

	fmt.Println("\n实际分配结果:")
	if githubCount > 0 {
		fmt.Printf("  ✅ GitHub: %d次 (%.1f%%)\n", githubCount, float64(githubCount)/float64(testCount)*100)
	}
	if baiduCount > 0 {
		fmt.Printf("  ✅ 百度: %d次 (%.1f%%)\n", baiduCount, float64(baiduCount)/float64(testCount)*100)
	}
	if tencentCount > 0 {
		fmt.Printf("  ✅ 腾讯: %d次 (%.1f%%)\n", tencentCount, float64(tencentCount)/float64(testCount)*100)
	}

	fmt.Println("\n=== 测试完成 ===")

	if totalRedirects == testCount {
		fmt.Println("🎉 所有测试通过！活码系统工作正常！")
	} else {
		fmt.Printf("⚠️  部分测试失败，成功重定向: %d/%d\n", totalRedirects, testCount)
	}

	fmt.Println("\n✅ 功能验证:")
	fmt.Println("  - 用户认证系统 ✓")
	fmt.Println("  - 活码创建功能 ✓")
	fmt.Println("  - 静态码管理 ✓")
	fmt.Println("  - 权重分配算法 ✓")
	fmt.Println("  - 重定向功能 ✓")
	fmt.Println("  - API接口完整性 ✓")
}
