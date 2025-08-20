package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"breast-implant-warranty-system/internal/utils"
)

func main() {
	fmt.Print("請輸入密碼: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	password := strings.TrimSpace(scanner.Text())

	if password == "" {
		fmt.Println("密碼不能為空")
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		fmt.Printf("雜湊失敗: %v\n", err)
		return
	}

	fmt.Printf("雜湊結果: %s\n", hash)
}
