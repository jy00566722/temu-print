package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// LabelData 定义标签数据
type LabelData struct {
	ServiceType   string // 物流名称
	PhoneNumber   string // 物流单号
	ItemNumber    string // 货号
	Quantity      int    // 商品数量
	TotalItems    int    // 总件数
	Warehouse     string // 店铺名称
	ShippingCrate string // 收货仓
	CurrentTime   string // 时间戳
}

// 定义标签尺寸（10cm x 10cm，转为毫米）
const (
	PageWidth  = 100.0 // mm
	PageHeight = 100.0 // mm
)

func main() {
	// 定义命令行参数
	outputPath := flag.String("output", "logistics_label.pdf", "输出PDF文件路径")
	autoPrint := flag.Bool("print", true, "生成PDF后是否自动打印")
	flag.Parse()

	// 使用示例数据，这些可以通过参数传入
	data := LabelData{
		ServiceType:   "邮政特快专递",
		PhoneNumber:   "13616578186xx",
		ItemNumber:    "8559",
		Quantity:      8,
		TotalItems:    1,
		Warehouse:     "建闽店",
		ShippingCrate: "三水一产25号子仓",
		CurrentTime:   "", // 会在生成时填充
	}

	// 将当前时间格式化为中文
	data.CurrentTime = formatChineseDateTime(time.Now())

	// 生成PDF文件
	pdfPath := generateLogisticsLabel(data, *outputPath)
	fmt.Printf("标签已生成: %s\n", pdfPath)

	// 如果需要自动打印
	if *autoPrint {
		err := printPDF(pdfPath)
		if err != nil {
			fmt.Printf("打印失败: %v\n", err)
		} else {
			fmt.Println("标签已发送到打印机")
		}
	}
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
		t.Year(), t.Month(), t.Day(), weekday, ampm, hour, t.Minute())
}

// generateLogisticsLabel 生成物流标签PDF文件
func generateLogisticsLabel(data LabelData, outputPath string) string {
	// 创建PDF文档，度量单位为毫米，页面大小为100mm x 100mm
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        "custom",
		Size:           gofpdf.SizeType{Wd: PageWidth, Ht: PageHeight},
		FontDirStr:     "",
	})
	pdf.AddPage()

	// 加载中文字体 - 需要设置适当的字体路径
	// 确保这些字体文件存在于指定路径，或者修改为系统中的字体路径
	fontPath := "fonts"
	err := os.MkdirAll(fontPath, 0755)
	if err != nil {
		fmt.Printf("创建字体目录失败: %v\n", err)
	}

	// 检查是否已经存在字体文件，如果不存在，给出提示
	regularFont := filepath.Join(fontPath, "NotoSansSC-Regular.ttf")
	boldFont := filepath.Join(fontPath, "NotoSansSC-Bold.ttf")

	if _, err := os.Stat(regularFont); os.IsNotExist(err) {
		fmt.Printf("警告: 字体文件 %s 不存在，可能导致中文显示问题\n", regularFont)
		fmt.Println("请从 https://fonts.google.com/noto/specimen/Noto+Sans+SC 下载字体文件")

		// 使用内置字体作为替代
		pdf.SetFont("Arial", "", 12)
	} else {
		pdf.AddUTF8Font("NotoSansSC", "", regularFont)
		pdf.AddUTF8Font("NotoSansSC", "B", boldFont)
	}

	// 设置外边框
	pdf.SetLineWidth(0.5)
	pdf.Rect(5, 5, 90, 90, "D")

	// 第一部分：标题和联系方式（顶部黑底白字部分）
	pdf.SetFillColor(0, 0, 0) // 黑色背景
	pdf.Rect(30, 8, 40, 10, "F")

	pdf.SetFont("NotoSansSC", "B", 16)
	pdf.SetTextColor(255, 255, 255) // 白色文本
	pdf.Text(32, 15, "TEMU物流单")

	pdf.SetFont("NotoSansSC", "B", 16)
	pdf.SetTextColor(0, 0, 0) // 黑色文本
	pdf.Text(15, 25, data.ServiceType)

	pdf.SetFont("NotoSansSC", "", 22)
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
	pdf.SetFont("NotoSansSC", "B", 18)
	pdf.Text(20, 57, data.ItemNumber)
	pdf.Text(70, 57, strconv.Itoa(data.Quantity))

	// pdf.SetFont("NotoSansSC", "", 18)
	// pdf.Text(20, 87, data.Description)

	// 第三部分：件数
	pdf.SetFont("NotoSansSC", "B", 16)
	pdf.Text(8, 76, fmt.Sprintf("共%d件", data.TotalItems))
	// 店铺
	pdf.SetFont("NotoSansSC", "", 16)
	pdf.Text(74, 76, data.Warehouse)

	// 收货仓信息和时间戳
	pdf.SetFont("NotoSansSC", "", 14)
	pdf.Text(12, 86, fmt.Sprintf("收货仓: %s", data.ShippingCrate))
	pdf.Text(10, 93, data.CurrentTime)

	// 保存PDF文件
	// 如果提供的是相对路径，则转换为绝对路径
	if !filepath.IsAbs(outputPath) {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("获取当前目录失败: %v\n", err)
			outputPath = filepath.Join(os.TempDir(), outputPath)
		} else {
			outputPath = filepath.Join(dir, outputPath)
		}
	}

	// 确保目录存在
	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		outputPath = filepath.Join(os.TempDir(), "logistics_label.pdf")
	}

	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		fmt.Printf("保存PDF文件失败: %v\n", err)
		outputPath = filepath.Join(os.TempDir(), "logistics_label.pdf")
		err = pdf.OutputFileAndClose(outputPath)
		if err != nil {
			panic(fmt.Sprintf("无法生成PDF文件: %v", err))
		}
	}

	return outputPath
}

// printPDF 调用系统打印机打印PDF文件
func printPDF(pdfPath string) error {
	// 针对Windows系统
	if runtime.GOOS == "windows" {
		// 方法1：使用SumatraPDF进行静默打印（需要安装）
		sumatra := "SumatraPDF.exe"
		cmd := exec.Command(sumatra, "-print-to-default", "-silent", pdfPath)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("使用SumatraPDF打印失败: %v\n", err)
			fmt.Println("尝试使用Adobe Reader进行打印...")

			// 方法2：使用Adobe Reader进行打印（需要安装）
			adobeReader := "AcroRd32.exe"
			cmd = exec.Command(adobeReader, "/t", pdfPath)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("使用Adobe Reader打印失败: %v\n", err)
				fmt.Println("尝试使用系统默认打印命令...")

				// 方法3：使用系统默认打印命令
				cmd = exec.Command("rundll32.exe", "mshtml.dll,PrintHTML", pdfPath)
				return cmd.Run()
			}
		}
		return err
	} else if runtime.GOOS == "linux" {
		// Linux系统下使用CUPS打印
		cmd := exec.Command("lpr", pdfPath)
		return cmd.Run()
	} else if runtime.GOOS == "darwin" {
		// macOS系统下使用lpr打印
		cmd := exec.Command("lpr", pdfPath)
		return cmd.Run()
	}

	return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
}
