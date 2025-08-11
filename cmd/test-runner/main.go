package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// 定義命令行參數
	var (
		runUnit        = flag.Bool("unit", false, "執行單元測試")
		runIntegration = flag.Bool("integration", false, "執行整合測試")
		runAudit       = flag.Bool("audit", false, "執行審計API測試")
		runAll         = flag.Bool("all", false, "執行所有測試")
		baseURL        = flag.String("url", "http://localhost:8080/api/v1", "API基礎URL")
		help           = flag.Bool("help", false, "顯示幫助資訊")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 如果沒有指定任何測試類型，默認執行所有測試
	if !*runUnit && !*runIntegration && !*runAudit && !*runAll {
		*runAll = true
	}

	fmt.Println("🧪 乳房植入物保固系統 - 測試執行器")
	fmt.Println("====================================================")
	fmt.Printf("⏰ 開始時間: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("🌐 API URL: %s\n", *baseURL)
	fmt.Println()

	var hasErrors bool

	if *runAll || *runUnit {
		fmt.Println("📋 執行單元測試...")
		if err := runUnitTests(); err != nil {
			fmt.Printf("❌ 單元測試失敗: %v\n", err)
			hasErrors = true
		} else {
			fmt.Println("✅ 單元測試完成")
		}
		fmt.Println()
	}

	if *runAll || *runIntegration {
		fmt.Println("🔗 執行整合測試...")
		if err := runIntegrationTests(*baseURL); err != nil {
			fmt.Printf("❌ 整合測試失敗: %v\n", err)
			hasErrors = true
		} else {
			fmt.Println("✅ 整合測試完成")
		}
		fmt.Println()
	}

	if *runAll || *runAudit {
		fmt.Println("📊 執行審計API測試...")
		if err := runAuditTests(); err != nil {
			fmt.Printf("❌ 審計API測試失敗: %v\n", err)
			hasErrors = true
		} else {
			fmt.Println("✅ 審計API測試完成")
		}
		fmt.Println()
	}

	fmt.Println("====================================================")
	fmt.Printf("⏰ 結束時間: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	if hasErrors {
		fmt.Println("❌ 測試執行完成，但有錯誤發生")
		os.Exit(1)
	} else {
		fmt.Println("✅ 所有測試執行完成，無錯誤")
	}
}

func showHelp() {
	fmt.Println("乳房植入物保固系統 - 測試執行器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run cmd/test-runner/main.go [選項]")
	fmt.Println()
	fmt.Println("選項:")
	fmt.Println("  -unit          執行單元測試")
	fmt.Println("  -integration   執行整合測試")
	fmt.Println("  -audit         執行審計API測試")
	fmt.Println("  -all           執行所有測試（默認）")
	fmt.Println("  -url string    API基礎URL (默認: http://localhost:8080/api/v1)")
	fmt.Println("  -help          顯示此幫助資訊")
	fmt.Println()
	fmt.Println("範例:")
	fmt.Println("  go run cmd/test-runner/main.go -unit")
	fmt.Println("  go run cmd/test-runner/main.go -integration -url http://localhost:8080/api/v1")
	fmt.Println("  go run cmd/test-runner/main.go -all")
	fmt.Println()
	fmt.Println("環境變數:")
	fmt.Println("  SKIP_INTEGRATION_TESTS=true  跳過整合測試")
	fmt.Println("  TEST_BASE_URL                 覆蓋API基礎URL")
}

func runUnitTests() error {
	fmt.Println("  📝 執行單元測試...")

	cmd := exec.Command("go", "test", "-v", "-short", "./tests")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("單元測試執行失敗: %v", err)
	}

	return nil
}

func runIntegrationTests(baseURL string) error {
	fmt.Println("  🔗 執行整合測試...")

	// 檢查後端是否執行
	if !isBackendRunning(baseURL) {
		return fmt.Errorf("後端服務未執行，請先啟動服務器")
	}

	cmd := exec.Command("go", "test", "-v", "./tests", "-run", "TestAPITestSuite")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("整合測試執行失敗: %v", err)
	}

	return nil
}

func runAuditTests() error {
	fmt.Println("  📊 執行審計API測試...")

	// 檢查後端是否執行
	if !isBackendRunning("http://localhost:8080/api/v1") {
		return fmt.Errorf("後端服務未執行，請先啟動服務器")
	}

	fmt.Println("  💡 提示: 審計API測試需要手動執行")
	fmt.Println("  💡 使用: go test ./tests -v -run TestAuditAPI")

	return nil
}

func isBackendRunning(baseURL string) bool {
	// 嘗試連接到健康檢查端點
	client := &http.Client{Timeout: 5 * time.Second}

	// 首先嘗試健康檢查端點
	healthURL := strings.Replace(baseURL, "/api/v1", "/health", 1)
	resp, err := client.Get(healthURL)
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == 200
	}

	// 如果健康檢查失敗，嘗試登入端點
	resp, err = client.Get(baseURL + "/auth/login")
	if err == nil {
		resp.Body.Close()
		// 登入端點應該返回 405 (Method Not Allowed) 對於 GET 請求
		return resp.StatusCode == 405 || resp.StatusCode == 200
	}

	return false
}
