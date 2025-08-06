package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"audio-loss-checker/internal/decoder"
	"audio-loss-checker/internal/types"

	"github.com/schollz/progressbar/v3"
)

// Analyzer 音频分析器
type Analyzer struct {
	config          *types.AnalyzerConfig
	decoderRegistry *decoder.DecoderRegistry
}

// NewAnalyzer 创建新的分析器
func NewAnalyzer(config *types.AnalyzerConfig) *Analyzer {
	return &Analyzer{
		config:          config,
		decoderRegistry: decoder.NewDecoderRegistry(),
	}
}

// AnalyzeFiles 分析多个音频文件
func (a *Analyzer) AnalyzeFiles(filePaths []string) error {
	// 创建进度条
	var bar *progressbar.ProgressBar
	if !a.config.Quiet && !a.config.JSONOutput {
		bar = progressbar.NewOptions(len(filePaths),
			progressbar.OptionSetDescription("分析音频文件"),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(50),
			progressbar.OptionShowIts(),
		)
	}

	// 创建工作通道
	jobs := make(chan string, len(filePaths))
	results := make(chan *types.AnalysisResult, len(filePaths))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < a.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				result := a.analyzeFile(filePath)
				results <- result
				if bar != nil {
					bar.Add(1)
				}
			}
		}()
	}

	// 发送任务
	go func() {
		for _, filePath := range filePaths {
			jobs <- filePath
		}
		close(jobs)
	}()

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集并输出结果
	var allResults []*types.AnalysisResult
	for result := range results {
		allResults = append(allResults, result)
		a.outputResult(result)
	}

	if bar != nil {
		bar.Finish()
		fmt.Println() // 换行
	}

	// 输出统计信息
	if !a.config.Quiet && !a.config.JSONOutput {
		a.printSummary(allResults)
	}

	return nil
}

// analyzeFile 分析单个音频文件
func (a *Analyzer) analyzeFile(filePath string) *types.AnalysisResult {
	result := &types.AnalysisResult{
		FilePath: filePath,
		Status:   "ERROR",
	}

	// 解码音频文件
	audioFile, err := a.decoderRegistry.DecodeFile(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("解码失败: %v", err)
		return result
	}
	defer audioFile.Close()

	// 填充基本信息
	result.Format = audioFile.GetFormat()
	result.Metadata = audioFile.GetMetadata()

	// 获取音频采样数据
	samples, err := audioFile.GetSamples()
	if err != nil {
		result.Error = fmt.Sprintf("读取音频数据失败: %v", err)
		return result
	}

	// 创建频谱分析器
	spectrumAnalyzer := NewSpectrumAnalyzer(audioFile.GetSampleRate())

	// 进行频谱分析
	spectrumResult, err := spectrumAnalyzer.AnalyzeSpectrum(samples)
	if err != nil {
		result.Error = fmt.Sprintf("频谱分析失败: %v", err)
		return result
	}

	// 填充分析结果
	result.Analysis = types.AnalysisDetails{
		IsFake:       spectrumResult.IsFake,
		CutoffHz:     spectrumResult.CutoffFrequency,
		Details:      spectrumResult.Details,
		SampleRate:   audioFile.GetSampleRate(),
		BitDepth:     audioFile.GetBitDepth(),
		Channels:     audioFile.GetChannels(),
		Duration:     audioFile.GetDuration().Seconds(),
		MaxFrequency: spectrumResult.MaxFrequency,
	}

	// 根据自定义截断频率判断
	if spectrumResult.MaxFrequency < a.config.CutoffFreq {
		result.Analysis.IsFake = true
		if !spectrumResult.IsFake {
			result.Analysis.Details = fmt.Sprintf("最高频率 %.0f Hz 低于设定阈值 %.0f Hz",
				spectrumResult.MaxFrequency, a.config.CutoffFreq)
		}
	}

	// 设置状态
	if result.Analysis.IsFake {
		result.Status = "FAKE"
	} else {
		result.Status = "OK"
	}

	return result
}

// outputResult 输出单个分析结果
func (a *Analyzer) outputResult(result *types.AnalysisResult) {
	// 如果只显示假无损文件，跳过正常文件
	if a.config.OnlyFake && result.Status != "FAKE" {
		return
	}

	// 静默模式，只输出假无损文件路径
	if a.config.Quiet {
		if result.Status == "FAKE" {
			fmt.Println(result.FilePath)
		}
		return
	}

	// JSON输出格式
	if a.config.JSONOutput {
		jsonData, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON序列化失败: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// 普通格式输出
	a.printDetailedResult(result)
}

// printDetailedResult 打印详细结果
func (a *Analyzer) printDetailedResult(result *types.AnalysisResult) {
	fmt.Printf("\n=== %s ===\n", filepath.Base(result.FilePath))
	fmt.Printf("路径: %s\n", result.FilePath)
	fmt.Printf("格式: %s\n", result.Format)
	fmt.Printf("状态: %s\n", result.Status)

	if result.Error != "" {
		fmt.Printf("错误: %s\n", result.Error)
		return
	}

	// 基本信息
	fmt.Printf("采样率: %d Hz\n", result.Analysis.SampleRate)
	fmt.Printf("位深度: %d bit\n", result.Analysis.BitDepth)
	fmt.Printf("声道数: %d\n", result.Analysis.Channels)
	fmt.Printf("时长: %.2f 秒\n", result.Analysis.Duration)

	// 元数据
	if result.Metadata.Title != "" {
		fmt.Printf("标题: %s\n", result.Metadata.Title)
	}
	if result.Metadata.Artist != "" {
		fmt.Printf("艺术家: %s\n", result.Metadata.Artist)
	}
	if result.Metadata.Album != "" {
		fmt.Printf("专辑: %s\n", result.Metadata.Album)
	}

	// 频谱分析结果
	fmt.Printf("最高有效频率: %.0f Hz\n", result.Analysis.MaxFrequency)
	if result.Analysis.CutoffHz > 0 {
		fmt.Printf("截断频率: %.0f Hz\n", result.Analysis.CutoffHz)
	}
	fmt.Printf("分析结果: %s\n", result.Analysis.Details)

	// 如果是假无损，用红色标记
	if result.Analysis.IsFake {
		fmt.Printf("⚠️  警告: 这可能是一个假无损文件！\n")
	} else {
		fmt.Printf("✅ 文件看起来是真实的无损音频\n")
	}
}

// printSummary 打印统计摘要
func (a *Analyzer) printSummary(results []*types.AnalysisResult) {
	total := len(results)
	fake := 0
	ok := 0
	errors := 0

	for _, result := range results {
		switch result.Status {
		case "FAKE":
			fake++
		case "OK":
			ok++
		case "ERROR":
			errors++
		}
	}

	fmt.Printf("\n=== 分析统计 ===\n")
	fmt.Printf("总文件数: %d\n", total)
	fmt.Printf("正常文件: %d\n", ok)
	fmt.Printf("假无损文件: %d\n", fake)
	if errors > 0 {
		fmt.Printf("错误文件: %d\n", errors)
	}

	if fake > 0 {
		fmt.Printf("\n⚠️  发现 %d 个可疑的假无损文件，建议进一步检查！\n", fake)
	} else {
		fmt.Printf("\n✅ 所有文件都看起来是真实的无损音频\n")
	}
}
