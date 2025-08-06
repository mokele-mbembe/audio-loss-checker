package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"audio-loss-checker/internal/analyzer"
	"audio-loss-checker/internal/types"

	"github.com/spf13/cobra"
)

var (
	quiet       bool
	onlyFake    bool
	jsonOutput  bool
	cutoffFreq  float64
	concurrency int
	version     = "1.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "audio-loss-checker [path]",
	Short: "检测无损音频文件是否真的是无损格式",
	Long: `Audio Loss Checker 是一个CLI工具，用于检测无损音频文件是否真的是无损格式。
当前支持 WAV, FLAC 格式。ALAC 和 APE 格式支持开发中。

通过频谱分析检测音频文件是否存在高频截断，从而判断是否为从有损格式转换而来的"假无损"文件。`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalysis,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "静默模式，仅输出假无损文件路径")
	rootCmd.Flags().BoolVar(&onlyFake, "only-fake", false, "只显示假无损文件的分析报告")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "以JSON格式输出结果")
	rootCmd.Flags().Float64Var(&cutoffFreq, "cutoff", 18000, "频率截断阈值 (Hz)")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "j", runtime.NumCPU(), "并发处理文件数量")
	rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")

	// 添加版本命令
	rootCmd.SetVersionTemplate("audio-loss-checker version {{.Version}}\n")
	rootCmd.Version = version
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	targetPath := args[0]

	// 检查路径是否存在
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", targetPath)
	}

	// 创建分析器配置
	config := &types.AnalyzerConfig{
		CutoffFreq:  cutoffFreq,
		Concurrency: concurrency,
		Quiet:       quiet,
		OnlyFake:    onlyFake,
		JSONOutput:  jsonOutput,
	}

	// 创建分析器实例
	audioAnalyzer := analyzer.NewAnalyzer(config)

	// 收集音频文件
	files, err := collectAudioFiles(targetPath)
	if err != nil {
		return fmt.Errorf("收集音频文件失败: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("未找到支持的音频文件")
		return nil
	}

	// 开始分析
	return audioAnalyzer.AnalyzeFiles(files)
}

func collectAudioFiles(path string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{
		".wav":  true,
		".flac": true,
		// TODO: 待实现的格式
		// ".alac": true,
		// ".ape":  true,
		// ".m4a":  true, // ALAC files often use .m4a extension
	}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(strings.ToLower(filePath))
		if supportedExts[ext] {
			files = append(files, filePath)
		}

		return nil
	})

	return files, err
}
