package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx   context.Context
	fonts embed.FS // 添加这一行，用于存储嵌入的字体
}

// LabelData 定义标签数据
type LabelData struct {
	ServiceType   string `json:"serviceType"`   // 物流名称
	PhoneNumber   string `json:"phoneNumber"`   // 物流单号
	ItemNumber    string `json:"itemNumber"`    // 货号
	Quantity      int    `json:"quantity"`      // 商品数量
	TotalItems    int    `json:"totalItems"`    // 总件数
	Warehouse     string `json:"warehouse"`     // 店铺名称
	ShippingCrate string `json:"shippingCrate"` // 收货仓
	CurrentTime   string `json:"currentTime"`   // 时间戳
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 在启动时获取嵌入的字体
	a.fonts = fonts
}

// ParseLogisticsInfo 解析物流信息文本
func (a *App) ParseLogisticsInfo(text string) map[string]interface{} {
	result := make(map[string]interface{})

	// 提取物流名称和单号
	logisticsPattern := regexp.MustCompile(`(?:物流单号：|邮政|快递|物流).*?[,，]?\s*(\S+)[,，]\s*(\d+)`)
	if matches := logisticsPattern.FindStringSubmatch(text); len(matches) >= 3 {
		result["serviceType"] = matches[1]
		result["phoneNumber"] = matches[2]
	}

	// 提取货号
	itemNumberPattern := regexp.MustCompile(`货号[：:]\s*(\d+)`)
	if matches := itemNumberPattern.FindStringSubmatch(text); len(matches) >= 2 {
		result["itemNumber"] = matches[1]
	}

	// 提取商品数量
	quantityPattern := regexp.MustCompile(`发货数量[：:]\s*(\d+)`)
	if matches := quantityPattern.FindStringSubmatch(text); len(matches) >= 2 {
		quantity, _ := strconv.Atoi(matches[1])
		result["quantity"] = quantity
	}

	// 提取收货仓
	shippingCratePattern := regexp.MustCompile(`收货仓库[：:]\s*(.+?)\s*(?:$|\n)`)
	if matches := shippingCratePattern.FindStringSubmatch(text); len(matches) >= 2 {
		result["shippingCrate"] = matches[1]
	}

	return result
}

// GenerateAndPrintLabel 生成物流标签并打印
func (a *App) GenerateAndPrintLabel(data LabelData, autoPrint bool) string {
	// 设置当前时间
	data.CurrentTime = formatChineseDateTime(time.Now())

	// 生成PDF文件
	outputPath := "logistics_label.pdf"
	pdfPath := a.generateLogisticsLabel(data, outputPath)

	// 如果需要自动打印
	message := fmt.Sprintf("标签已生成: %s", pdfPath)
	if autoPrint {
		err := a.printPDF(pdfPath)
		if err != nil {
			message = fmt.Sprintf("标签已生成，但打印失败: %v", err)
		} else {
			message = "标签已生成并发送到打印机"
		}
	}

	return message
}

// formatChineseDateTime 格式化日期时间为中文格式
func formatChineseDateTime(t time.Time) string {
	weekdays := []string{"日", "一", "二", "三", "四", "五", "六"}
	weekday := weekdays[t.Weekday()]

	hour := t.Hour()
	ampm := "上午"
	if hour >= 12 {
		ampm = "下午"
		if hour > 12 {
			hour -= 12
		}
	}

	return fmt.Sprintf("%d年%d月%d日, 星期%s %s %d:%02d",
		t.Year(), int(t.Month()), t.Day(), weekday, ampm, hour, t.Minute())
}

// generateLogisticsLabel 生成物流标签PDF文件
func (a *App) generateLogisticsLabel(data LabelData, outputPath string) string {
	// 定义标签尺寸（10cm x 10cm，转为毫米）
	const (
		PageWidth  = 100.0 // mm
		PageHeight = 100.0 // mm
	)

	// 创建PDF文档，度量单位为毫米，页面大小为100mm x 100mm
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        "custom",
		Size:           gofpdf.SizeType{Wd: PageWidth, Ht: PageHeight},
		FontDirStr:     "",
	})

	// 添加嵌入的中文字体
	// 1. 从嵌入资源读取字体文件
	regularFontBytes, err := a.fonts.ReadFile("fonts/NotoSansSC-Regular.ttf")
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("读取常规字体文件失败: %v", err))
	}

	boldFontBytes, err := a.fonts.ReadFile("fonts/NotoSansSC-Bold.ttf")
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("读取粗体字体文件失败: %v", err))
	}

	// 2. 添加字体到PDF
	pdf.AddUTF8FontFromBytes("NotoSansSC", "", regularFontBytes)

	pdf.AddUTF8FontFromBytes("NotoSansSC", "B", boldFontBytes)

	pdf.AddPage()

	// 使用添加的中文字体
	pdf.SetFont("NotoSansSC", "", 12)

	// 设置外边框
	pdf.SetLineWidth(0.5)
	pdf.Rect(5, 5, 90, 90, "D")

	// 第一部分：标题和联系方式（顶部黑底白字部分）
	pdf.SetFillColor(0, 0, 0) // 黑色背景
	pdf.Rect(30, 8, 40, 10, "F")

	pdf.SetFont("NotoSansSC", "B", 16) // 使用粗体
	pdf.SetTextColor(255, 255, 255)    // 白色文本
	pdf.Text(32, 15, "TEMU物流单")

	pdf.SetFont("NotoSansSC", "B", 16) // 使用粗体
	pdf.SetTextColor(0, 0, 0)          // 黑色文本
	pdf.Text(15, 25, data.ServiceType)

	pdf.SetFont("NotoSansSC", "", 22) // 普通样式
	pdf.Text(25, 35, data.PhoneNumber)

	// 第二部分：货物信息表格
	pdf.Line(5, 40, 95, 40)  // 横线
	pdf.Line(50, 40, 50, 80) // 竖线
	pdf.Line(5, 50, 95, 50)  // 第二横线
	pdf.Line(5, 60, 95, 60)  // 第三横线
	pdf.Line(5, 70, 95, 70)  // 第四横线
	pdf.Line(5, 80, 95, 80)  // 底线

	// 表格标题
	pdf.SetFont("NotoSansSC", "", 16)
	pdf.Text(20, 46, "货号")
	pdf.Text(70, 46, "数量/双")

	// 表格内容
	pdf.SetFont("NotoSansSC", "B", 18) // 使用粗体
	pdf.Text(20, 57, data.ItemNumber)
	pdf.Text(70, 57, strconv.Itoa(data.Quantity))

	// 第三部分：件数
	pdf.SetFont("NotoSansSC", "B", 16) // 使用粗体
	pdf.Text(8, 76, fmt.Sprintf("共%d件", data.TotalItems))
	// 店铺
	pdf.SetFont("NotoSansSC", "", 16)
	pdf.Text(74, 76, data.Warehouse)

	// 收货仓信息和时间戳
	pdf.SetFont("NotoSansSC", "", 14)
	pdf.Text(12, 86, fmt.Sprintf("收货仓: %s", data.ShippingCrate))
	pdf.Text(10, 93, data.CurrentTime)

	// 生成文件路径和保存逻辑保持不变
	homeDir, err := os.UserHomeDir()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("获取用户主目录失败: %v", err))
		return filepath.Join(os.TempDir(), outputPath)
	}

	// 在 Documents 目录下创建应用专用文件夹
	appFolder := filepath.Join(homeDir, "Documents", "TEMU-Labels")
	err = os.MkdirAll(appFolder, 0755)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("创建应用目录失败: %v", err))
		return filepath.Join(os.TempDir(), outputPath)
	}

	// 生成带时间戳的文件名，避免文件覆盖
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("logistics_label_%s.pdf", timestamp)
	outputPath = filepath.Join(appFolder, fileName)

	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("保存PDF文件失败: %v", err))
		// 尝试保存到临时目录
		outputPath = filepath.Join(os.TempDir(), "logistics_label.pdf")
		err = pdf.OutputFileAndClose(outputPath)
		if err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("无法生成PDF文件: %v", err))
		}
	}

	return outputPath
}

// printPDF 调用系统打印机打印PDF文件
func (a *App) printPDF(pdfPath string) error {
	// 针对Windows系统
	if os.Getenv("GOOS") == "windows" {
		// 方法1：使用SumatraPDF进行静默打印（需要安装）
		sumatra := "SumatraPDF.exe"
		cmd := exec.Command(sumatra, "-print-to-default", "-silent", pdfPath)
		err := cmd.Run()
		if err != nil {
			runtime.LogWarning(a.ctx, fmt.Sprintf("使用SumatraPDF打印失败: %v", err))
			runtime.LogInfo(a.ctx, "尝试使用Adobe Reader进行打印...")

			// 方法2：使用Adobe Reader进行打印（需要安装）
			adobeReader := "AcroRd32.exe"
			cmd = exec.Command(adobeReader, "/t", pdfPath)
			err = cmd.Run()
			if err != nil {
				runtime.LogWarning(a.ctx, fmt.Sprintf("使用Adobe Reader打印失败: %v", err))
				runtime.LogInfo(a.ctx, "尝试使用系统默认打印命令...")

				// 方法3：使用系统默认打印命令
				cmd = exec.Command("rundll32.exe", "mshtml.dll,PrintHTML", pdfPath)
				return cmd.Run()
			}
		}
		return err
	} else if os.Getenv("GOOS") == "linux" {
		// Linux系统下使用CUPS打印
		cmd := exec.Command("lpr", pdfPath)
		return cmd.Run()
	} else if os.Getenv("GOOS") == "darwin" {
		// macOS系统下使用lpr打印
		cmd := exec.Command("lpr", pdfPath)
		return cmd.Run()
	}

	return fmt.Errorf("不支持的操作系统: %s", os.Getenv("GOOS"))
}

// OpenPDF 打开生成的PDF文件
func (a *App) OpenPDF(pdfPath string) {
	// 使用 runtime.Browser 打开 PDF 文件
	runtime.BrowserOpenURL(a.ctx, "file://"+pdfPath)

}
