package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PANGenerator struct {
	usedPANs map[string]bool
	mutex    sync.RWMutex
}

func NewPANGenerator() *PANGenerator {
	return &PANGenerator{
		usedPANs: make(map[string]bool),
		mutex:    sync.RWMutex{},
	}
}

// Tính toán check digit theo thuật toán Luhn
func calculateLuhnCheckDigit(digits []int) int {
	sum := 0
	isEven := true // Bắt đầu từ vị trí cuối (không tính check digit)

	// Duyệt từ phải sang trái (không tính check digit)
	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]

		if isEven {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		isEven = !isEven
	}

	// Check digit là số cần thêm để tổng chia hết cho 10
	return (10 - (sum % 10)) % 10
}

// Kiểm tra tính hợp lệ của PAN theo thuật toán Luhn
func validateLuhn(pan string) bool {
	if len(pan) != 16 {
		return false
	}

	sum := 0
	isEven := false

	// Duyệt từ phải sang trái
	for i := len(pan) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(pan[i]))
		if err != nil {
			return false
		}

		if isEven {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// Sinh số ngẫu nhiên an toàn
func generateSecureRandomNumber(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

// Sinh PAN 16 chữ số bắt đầu với "73"
func (pg *PANGenerator) GeneratePAN() (string, error) {
	pg.mutex.Lock()
	defer pg.mutex.Unlock()

	maxAttempts := 1000000 // Giới hạn số lần thử để tránh vòng lặp vô hạn

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Bắt đầu với "73"
		panDigits := []int{7, 3}

		// Sinh 13 chữ số tiếp theo (vị trí 3-15)
		for i := 0; i < 13; i++ {
			randomNum, err := generateSecureRandomNumber(10)
			if err != nil {
				return "", fmt.Errorf("lỗi sinh số ngẫu nhiên: %v", err)
			}
			panDigits = append(panDigits, int(randomNum))
		}

		// Tính check digit (vị trí 16)
		checkDigit := calculateLuhnCheckDigit(panDigits)
		panDigits = append(panDigits, checkDigit)

		// Chuyển thành string
		panStr := ""
		for _, digit := range panDigits {
			panStr += strconv.Itoa(digit)
		}

		// Kiểm tra trùng lặp
		if !pg.usedPANs[panStr] {
			pg.usedPANs[panStr] = true

			// Xác thực lại bằng thuật toán Luhn
			if validateLuhn(panStr) {
				return panStr, nil
			}
		}
	}

	return "", fmt.Errorf("không thể sinh PAN duy nhất sau %d lần thử", maxAttempts)
}

// Kiểm tra PAN đã được sử dụng chưa
func (pg *PANGenerator) IsPANUsed(pan string) bool {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	return pg.usedPANs[pan]
}

// Lấy số lượng PAN đã sinh
func (pg *PANGenerator) GetUsedCount() int {
	pg.mutex.RLock()
	defer pg.mutex.RUnlock()
	return len(pg.usedPANs)
}

// Format PAN để hiển thị (xxxx xxxx xxxx xxxx)
func formatPAN(pan string) string {
	if len(pan) != 16 {
		return pan
	}
	return fmt.Sprintf("%s %s %s %s => hash %s **** **** %s",
		pan[0:4], pan[4:8], pan[8:12], pan[12:16], pan[0:4], pan[12:16])
}

type CVVType int

const (
	CVV3Digit CVVType = 3 // Visa, MasterCard
	CVV4Digit CVVType = 4 // American Express
)

type CardInfo struct {
	PAN         string
	ExpiryDate  string
	ServiceCode string
}

type CVVGenerator struct {
	masterKey []byte
}

func NewCVVGenerator(key string) *CVVGenerator {
	hash := sha256.Sum256([]byte(key))
	return &CVVGenerator{
		masterKey: hash[:],
	}
}

func (cvv *CVVGenerator) GenerateDeterministicCVV(cardInfo CardInfo, cvvType CVVType) (string, error) {
	input := cardInfo.PAN + cardInfo.ExpiryDate + cardInfo.ServiceCode

	input += string(cvv.masterKey)

	hash := sha256.Sum256([]byte(input))

	hashStr := fmt.Sprintf("%x", hash)

	digits := int(cvvType)
	result := ""
	digitCount := 0

	for _, char := range hashStr {
		if digitCount >= digits {
			break
		}
		if char >= '0' && char <= '9' {
			result += string(char)
			digitCount++
		}
	}

	if len(result) < digits {
		for i := 0; i < len(hashStr) && len(result) < digits; i++ {
			char := hashStr[i]
			if char >= 'a' && char <= 'f' {
				digit := int(char-'a') % 10
				result += strconv.Itoa(digit)
			}
		}
	}

	if len(result) > digits {
		result = result[:digits]
	}

	return result, nil
}

func ValidateCVV(cvv string, cvvType CVVType) bool {
	if len(cvv) != int(cvvType) {
		return false
	}

	for _, char := range cvv {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func GetCVVTypeFromPAN(pan string) CVVType {
	if len(pan) >= 2 {
		prefix := pan[:2]
		if prefix == "34" || prefix == "37" {
			return CVV4Digit
		}
	}
	return CVV3Digit
}

func generateSampleCardInfo(pan string) CardInfo {
	mathrand.Seed(time.Now().UnixNano())
	now := time.Now()
	month := now.Month() + time.Month(mathrand.Intn(24))
	year := now.Year()
	if month > 12 {
		month = 6
		year += 6
	}

	expiryDate := fmt.Sprintf("%02d/%02d", int(month), year%100)

	serviceCodes := []string{"101", "121", "201", "221"}
	serviceCode := serviceCodes[mathrand.Intn(len(serviceCodes))]

	return CardInfo{
		PAN:         pan,
		ExpiryDate:  expiryDate,
		ServiceCode: serviceCode,
	}
}

func main() {
	generator := NewPANGenerator()

	fmt.Println("=== THUẬT TOÁN SINH PAN THẺ TÍN DỤNG ===")
	fmt.Println("Đầu số: 73xxxx")
	fmt.Println("Độ dài: 16 chữ số")
	fmt.Println("Thuật toán: Luhn Algorithm")
	fmt.Println("========================================")

	// Sinh và hiển thị 10 PAN mẫu
	for i := 1; i <= 10; i++ {
		pan, err := generator.GeneratePAN()
		if err != nil {
			fmt.Printf("Lỗi sinh PAN #%d: %v\n", i, err)
			continue
		}

		fmt.Printf("PAN #%02d: %s (%s)\n", i, pan, formatPAN(pan))

		// Xác thực
		if validateLuhn(pan) {
			fmt.Printf("         ✓ Hợp lệ theo Luhn Algorithm\n")
		} else {
			fmt.Printf("         ✗ Không hợp lệ theo Luhn Algorithm\n")
		}
		fmt.Println()
	}

	fmt.Printf("Tổng số PAN đã sinh: %d\n", generator.GetUsedCount())

	// Test kiểm tra trùng lặp
	fmt.Println("\n=== TEST KIỂM TRA TRÙNG LẶP ===")
	testPAN := "7312345678901234" // PAN test
	fmt.Printf("Kiểm tra PAN %s đã được sử dụng: %v\n",
		testPAN, generator.IsPANUsed(testPAN))

	fmt.Println("=== THUẬT TOÁN SINH CVV THẺ TÍN DỤNG ===")
	fmt.Println()

	cvvGenerator := NewCVVGenerator("MasterKey2024SecureBank")

	samplePANs := []string{
		"7312345678901234", // Custom card (3 digits)
	}

	for i, pan := range samplePANs {
		fmt.Printf("=== THẺ #%d ===\n", i+1)
		fmt.Printf("PAN: %s\n", pan)

		// Tạo thông tin thẻ
		cardInfo := generateSampleCardInfo(pan)
		fmt.Printf("Expiry: %s\n", cardInfo.ExpiryDate)
		fmt.Printf("Service Code: %s\n", cardInfo.ServiceCode)

		// Xác định loại CVV
		cvvType := GetCVVTypeFromPAN(pan)
		fmt.Printf("CVV Type: %d digits\n", int(cvvType))

		// Phương pháp 2: Deterministic CVV
		detCVV, err := cvvGenerator.GenerateDeterministicCVV(cardInfo, cvvType)
		if err != nil {
			fmt.Printf("Lỗi sinh Deterministic CVV: %v\n", err)
		} else {
			fmt.Printf("Deterministic CVV: %s\n", detCVV)
		}

		// Validation
		if ValidateCVV(detCVV, cvvType) {
			fmt.Printf("✓ Deterministic CVV hợp lệ\n")
		}

		fmt.Println(strings.Repeat("-", 40))
	}
}
