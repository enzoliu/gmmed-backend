package validator

import (
	"errors"
	"strings"
)

// https://en.wikipedia.org/wiki/Machine-readable_passport
func IsValidPassportNumber(passportNumber interface{}) error {
	p, ok := passportNumber.(string)
	if !ok {
		return errors.New("passport number must be a string")
	}

	CodeMap := map[rune]int{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
		'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'A': 10, 'B': 11, 'C': 12, 'D': 13,
		'E': 14, 'F': 15, 'G': 16, 'H': 17,
		'I': 18, 'J': 19, 'K': 20, 'L': 21,
		'M': 22, 'N': 23, 'O': 24, 'P': 25,
		'Q': 26, 'R': 27, 'S': 28, 'T': 29,
		'U': 30, 'V': 31, 'W': 32, 'X': 33,
		'Y': 34, 'Z': 35, '<': 0,
	}
	Weights := []int{7, 3, 1, 7, 3, 1, 7, 3, 1} // 權重因子

	if len(p) != 10 {
		return errors.New("passport number must be 10 characters long")
	}
	sum := 0
	for i, c := range strings.ToUpper(p) {
		if i == 9 {
			// 最後一位是檢查碼，不參與計算
			break
		}
		sum += CodeMap[c] * Weights[i]
	}
	checkDigit := sum % 10
	if checkDigit == int(p[9]-'0') {
		return nil // 檢查碼正確
	}
	return errors.New("invalid passport number: check digit does not match")
}

// https://zh.wikipedia.org/zh-tw/%E4%B8%AD%E8%8F%AF%E6%B0%91%E5%9C%8B%E5%9C%8B%E6%B0%91%E8%BA%AB%E5%88%86%E8%AD%89
func IsValidTaiwanID(taiwanID interface{}) error {
	id, ok := taiwanID.(string)
	if !ok {
		return errors.New("the Taiwan ID must be a string")
	}

	locMap := map[rune]int{
		'A': 1, 'B': 0, 'C': 9,
		'D': 8, 'E': 7, 'F': 6,
		'G': 5, 'H': 4, 'I': 9,
		'J': 3, 'K': 2, 'L': 2,
		'M': 1, 'N': 0, 'O': 8,
		'P': 9, 'Q': 8, 'R': 7,
		'S': 6, 'T': 5, 'U': 4,
		'V': 3, 'W': 1, 'X': 3,
		'Y': 2, 'Z': 0,
	}
	foreignOldCodeMap := map[rune]int{
		'A': 0, 'B': 1, 'C': 2, 'D': 3,
	}
	weights := []int{8, 7, 6, 5, 4, 3, 2, 1, 1} // 權重因子

	if len(id) != 10 {
		return errors.New("the Taiwan ID must be 10 characters long")
	}
	sum := 0
	for i, c := range strings.ToUpper(id) {
		if i == 0 {
			// 第一位是英文字母，轉換為數字
			if val, ok := locMap[c]; ok {
				sum += val
				continue
			} else {
				return errors.New("invalid first character in Taiwan ID")
			}
		} else if i == 1 {
			// 第二位可能會是居留證舊制的性別碼，必須是A、B、C或D
			if c >= 'A' && c <= 'D' {
				sum += foreignOldCodeMap[c] * weights[i-1]
				continue
			}
		} else if c < '0' || c > '9' {
			return errors.New("invalid character in Taiwan ID")
		}
		digitNum := int(c - '0')
		sum += digitNum * weights[i-1]
	}

	if sum%10 == 0 {
		return nil // 檢查碼正確
	}

	return errors.New("invalid Taiwan ID: check digit does not match")
}

func IsValidIdentity(identity interface{}) error {
	id, ok := identity.(string)
	if !ok {
		return errors.New("identity must be a string")
	}

	// 先檢查台灣身分證
	if err := IsValidTaiwanID(id); err == nil {
		return nil
	}
	// 如果不是台灣身分證，檢查護照號碼
	return IsValidPassportNumber(id)
}
