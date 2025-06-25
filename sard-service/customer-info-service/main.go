package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/unicode/norm"
)

// CCCDInfo struct để lưu thông tin CCCD
type CCCDInfo struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	DateOfBirth  string    `json:"date_of_birth"`
	Gender       string    `json:"gender"`
	Nationality  string    `json:"nationality"`
	PlaceOfBirth string    `json:"place_of_birth"`
	Address      string    `json:"address"`
	IssueDate    string    `json:"issue_date"`
	ExpiryDate   string    `json:"expiry_date"`
	IsDetected   bool      `json:"is_detected"`
	Confidence   float64   `json:"confidence"`
	ProcessedAt  time.Time `json:"processed_at"`
}

// Response struct cho API
type APIResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    CCCDInfo `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type ImageProcessor struct {
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func (ip *ImageProcessor) DetectCCCD(imageData []byte, filename string) (*CCCDInfo, error) {
	if !ip.isValidImageFormat(imageData) {
		return nil, fmt.Errorf("unsupported image format")
	}

	processedImageData, err := ip.preprocessImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %v", err)
	}

	ocrText, err := ip.performOCRFromBytesViaPipe(processedImageData)
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %v", err)
	}

	cccdInfo := ip.extractCCCDInfo(ocrText)
	cccdInfo.ProcessedAt = time.Now()

	return cccdInfo, nil
}

func (ip *ImageProcessor) isValidImageFormat(imageData []byte) bool {
	reader := bytes.NewReader(imageData)
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return false
	}

	validFormats := []string{"jpeg", "jpg", "png", "gif", "bmp"}
	for _, f := range validFormats {
		if strings.ToLower(format) == f {
			return true
		}
	}
	return false
}

func (ip *ImageProcessor) preprocessImage(imageData []byte) ([]byte, error) {
	reader := bytes.NewReader(imageData)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	processedImg := ip.enhanceImage(img)

	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(&buf, processedImg)
	default:
		err = png.Encode(&buf, processedImg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %v", err)
	}

	return buf.Bytes(), nil
}

func (ip *ImageProcessor) enhanceImage(img image.Image) image.Image {
	bounds := img.Bounds()

	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor)
			grayImg.Set(x, y, grayColor)
		}
	}

	enhancedImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := grayImg.GrayAt(x, y)
			newY := uint8(ip.clamp(int(gray.Y)*120/100, 0, 255))
			enhancedImg.SetGray(x, y, color.Gray{Y: newY})
		}
	}

	scaledImg := ip.scaleImage(enhancedImg, 2.0)

	return scaledImg
}

func (ip *ImageProcessor) scaleImage(img image.Image, scale float64) image.Image {
	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scale)
	newHeight := int(float64(bounds.Dy()) * scale)

	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)

			if srcX < bounds.Max.X && srcY < bounds.Max.Y {
				srcColor := img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y)
				scaledImg.Set(x, y, srcColor)
			}
		}
	}

	return scaledImg
}

func (ip *ImageProcessor) clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (ip *ImageProcessor) performOCRFromBytesViaPipe(imageData []byte) (string, error) {
	cmd := exec.Command("tesseract", "stdin", "stdout", "-l", "vie+eng", "--psm", "6")

	cmd.Stdin = bytes.NewReader(imageData)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract OCR failed: %v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

// removeDiacritics loại bỏ dấu tiếng Việt
func removeDiacritics(str string) string {
	t := norm.NFD.String(str)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

func cleanName(raw string) string {
	raw = strings.TrimSpace(raw)
	words := strings.Fields(raw)

	var nameWords []string
	for _, w := range words {
		if len(w) <= 1 {
			continue // loại từ 1 chữ cái như "Y", "U"
		}
		if matched, _ := regexp.MatchString(`^[A-ZÀÁẠẢÃÂẦẤẬẨẪĂẰẮẶẲẴÊỀẾỆỂỄĐÈÉẸẺẼÌÍỊỈĨÒÓỌỎÕÔỒỐỘỔỖƠỜỚỢỞỠÙÚỤỦŨƯỪỨỰỬỮỲÝỴỶỸ]+$`, w); matched {
			nameWords = append(nameWords, w)
		}
	}
	return strings.Join(nameWords, " ")
}

func (ip *ImageProcessor) extractCCCDInfo(ocrText string) *CCCDInfo {
	cccd := &CCCDInfo{
		IsDetected: false,
		Confidence: 0.0,
	}

	// Chuẩn hóa text
	raw := strings.ToUpper(ocrText)
	raw = regexp.MustCompile(`\s+`).ReplaceAllString(raw, " ")
	text := removeDiacritics(raw)

	// ==== Số CCCD ====
	idRegex := regexp.MustCompile(`(?:S[O0][\./:\-\s]*)?([0-9]{12})`)
	if match := idRegex.FindStringSubmatch(text); len(match) > 1 {
		cccd.ID = match[1]
		cccd.Confidence += 0.2
	}

	// ==== Họ tên ====
	fullNameRegex := regexp.MustCompile(`(?:HO VA TEN|FULL NAME|FULL NARNE)[\s:]*([A-ZÀÁẠẢÃÂĂĐÈÉẸẺẼÊÌÍỊỈĨÒÓỌỎÕÔƠÙÚỤỦƯỲÝỴỶỸ\s]{5,})`)
	if match := fullNameRegex.FindStringSubmatch(text); len(match) > 1 {
		cleaned := cleanName(match[1])
		if len(strings.Fields(cleaned)) >= 2 {
			cccd.FullName = cleaned
			cccd.Confidence += 0.2
		}
	}

	// ==== Ngày sinh ====
	dobLabelIndex := strings.Index(text, "NGAY SINH")
	if dobLabelIndex == -1 {
		dobLabelIndex = strings.Index(text, "DATE OF BIRTH")
	}
	dobRegex := regexp.MustCompile(`(\d{2}[\/\-]\d{2}[\/\-]\d{4})`)
	if matches := dobRegex.FindAllString(text, -1); len(matches) > 0 {
		cccd.DateOfBirth = strings.ReplaceAll(matches[0], "-", "/")
		cccd.Confidence += 0.1
	}

	// ==== Giới tính ====
	genderRegex := regexp.MustCompile(`(GIOI TINH|SEX)[^A-Z]*?(NAM|NU)`)
	if match := genderRegex.FindStringSubmatch(text); len(match) >= 3 {
		if match[2] == "NAM" {
			cccd.Gender = "Nam"
		} else if match[2] == "NU" {
			cccd.Gender = "Nữ"
		}
		cccd.Confidence += 0.05
	}

	// ==== Quốc tịch ====
	if strings.Contains(text, "VIET NAM") {
		cccd.Nationality = "Việt Nam"
		cccd.Confidence += 0.05
	}

	// ==== Quê quán ====
	originRegex := regexp.MustCompile(`(QUE QUAN|PLACE OF ORIGIN)[\s:/\-]*([A-ZÀÁẠẢÃÂĂĐÈÉẸẺẼÊÌÍỊỈĨÒÓỌỎÕÔƠÙÚỤỦƯỲÝỴỶỸ\s,]{5,})`)
	if match := originRegex.FindStringSubmatch(text); len(match) > 2 {
		place := strings.Trim(match[2], " ,.-:\n")
		if len(place) >= 5 {
			cccd.PlaceOfBirth = place
			cccd.Confidence += 0.1
		}
	}

	// ==== Nơi thường trú ====
	addressRegex := regexp.MustCompile(`(NOI THUONG TRU|PLACE OF RESIDENCE)[^\n]*[\n ]*([A-Z0-9ÀÁẠẢÃÂĂĐÈÉẸẺẼÊÌÍỊỈĨÒÓỌỎÕÔƠÙÚỤỦƯỲÝỴỶỸ,.'\-–—\s]{20,})`)
	if match := addressRegex.FindStringSubmatch(text); len(match) > 2 {
		addr := strings.Trim(match[2], " ,.-–—:\n")
		if len(addr) >= 10 {
			cccd.Address = addr
			cccd.Confidence += 0.1
		}
	}

	// ==== Ngày cấp & Ngày hết hạn ====
	datePattern := regexp.MustCompile(`\d{1,2}[\/\-]\d{1,2}[\/\-]\d{4}`)
	allDates := datePattern.FindAllString(text, -1)

	// Clean và loại bỏ DOB
	clean := func(s string) string { return strings.ReplaceAll(s, "-", "/") }
	dob := clean(cccd.DateOfBirth)
	dateMap := map[string]bool{}
	var validDates []string
	for _, d := range allDates {
		d = clean(d)
		if d != "" && d != dob && !dateMap[d] {
			dateMap[d] = true
			validDates = append(validDates, d)
		}
	}

	// Sắp xếp theo thời gian
	dates := []time.Time{}
	for _, d := range validDates {
		t, err := time.Parse("02/01/2006", d)
		if err == nil {
			dates = append(dates, t)
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	if len(dates) >= 2 {
		cccd.IssueDate = dates[0].Format("02/01/2006")
		cccd.ExpiryDate = dates[1].Format("02/01/2006")
		cccd.Confidence += 0.1
	} else if len(dates) == 1 {
		cccd.IssueDate = dates[0].Format("02/01/2006")
		cccd.Confidence += 0.05
	}

	// ==== Đánh dấu là phát hiện nếu có dữ liệu chính ====
	if cccd.ID != "" || cccd.FullName != "" || cccd.DateOfBirth != "" {
		cccd.IsDetected = true
	}

	return cccd
}

func setupRoutes(processor *ImageProcessor) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Endpoint chính để detect CCCD
	r.POST("/api/detect-cccd", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "No image file provided",
			})
			return
		}
		defer file.Close()

		// Đọc dữ liệu file
		imageData, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Failed to read image data",
			})
			return
		}

		// Kiểm tra kích thước file (max 10MB)
		if len(imageData) > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "File size too large (max 10MB)",
			})
			return
		}

		// Detect CCCD
		cccdInfo, err := processor.DetectCCCD(imageData, header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Detection failed: %v", err),
			})
			return
		}

		response := APIResponse{
			Success: true,
			Message: "Detection completed",
			Data:    *cccdInfo,
		}

		if !cccdInfo.IsDetected {
			response.Message = "No CCCD detected in image"
		}

		c.JSON(http.StatusOK, response)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
		})
	})

	// API info endpoint
	r.GET("/api/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "CCCD Detection API",
			"version":     "1.0.0",
			"description": "API để phát hiện và trích xuất thông tin CCCD từ ảnh",
			"endpoints": gin.H{
				"POST /api/detect-cccd": "Upload ảnh để detect CCCD",
				"GET /health":           "Kiểm tra trạng thái API",
				"GET /api/info":         "Thông tin API",
			},
		})
	})

	return r
}

func main() {
	if err := checkDependencies(); err != nil {
		log.Fatalf("Missing dependencies: %v", err)
	}

	processor := NewImageProcessor()

	router := setupRoutes(processor)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting CCCD Detection API on port %s", port)
	log.Printf("API endpoints:")
	log.Printf("  POST http://localhost:%s/api/detect-cccd", port)
	log.Printf("  GET  http://localhost:%s/health", port)
	log.Printf("  GET  http://localhost:%s/api/info", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func checkDependencies() error {
	// Kiểm tra Tesseract
	if _, err := exec.LookPath("tesseract"); err != nil {
		return fmt.Errorf("tesseract OCR not found - please install tesseract-ocr")
	}

	return nil
}
