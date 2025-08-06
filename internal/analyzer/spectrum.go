package analyzer

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
)

// SpectrumAnalyzer 频谱分析器
type SpectrumAnalyzer struct {
	sampleRate int
	windowSize int
}

// NewSpectrumAnalyzer 创建频谱分析器
func NewSpectrumAnalyzer(sampleRate int) *SpectrumAnalyzer {
	// 使用合适的窗口大小进行FFT分析
	windowSize := 8192 // 8K窗口，提供良好的频率分辨率
	return &SpectrumAnalyzer{
		sampleRate: sampleRate,
		windowSize: windowSize,
	}
}

// AnalyzeSpectrum 分析音频频谱
func (s *SpectrumAnalyzer) AnalyzeSpectrum(samples []float64) (*SpectrumResult, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("音频采样数据为空")
	}

	// 如果样本数量太少，使用所有样本
	if len(samples) < s.windowSize {
		s.windowSize = nearestPowerOf2(len(samples))
	}

	// 取样本的中间部分进行分析，避免开头和结尾的静音部分
	startIdx := len(samples) / 4
	endIdx := startIdx + s.windowSize
	if endIdx > len(samples) {
		endIdx = len(samples)
		startIdx = endIdx - s.windowSize
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// 提取分析窗口
	window := samples[startIdx:endIdx]

	// 应用汉明窗减少频谱泄漏
	windowedSamples := s.applyHammingWindow(window)

	// 进行FFT变换
	spectrum := fft.FFTReal(windowedSamples)

	// 计算功率谱密度
	powerSpectrum := s.calculatePowerSpectrum(spectrum)

	// 分析频谱特征
	result := s.analyzeFrequencyContent(powerSpectrum)

	return result, nil
}

// applyHammingWindow 应用汉明窗
func (s *SpectrumAnalyzer) applyHammingWindow(samples []float64) []float64 {
	windowed := make([]float64, len(samples))
	n := len(samples)

	for i, sample := range samples {
		// 汉明窗函数: w(n) = 0.54 - 0.46 * cos(2π * n / (N-1))
		window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		windowed[i] = sample * window
	}

	return windowed
}

// calculatePowerSpectrum 计算功率谱
func (s *SpectrumAnalyzer) calculatePowerSpectrum(spectrum []complex128) []float64 {
	power := make([]float64, len(spectrum)/2) // 只需要一半，因为FFT是对称的

	for i := 0; i < len(power); i++ {
		// 功率 = |复数|^2
		power[i] = cmplx.Abs(spectrum[i]) * cmplx.Abs(spectrum[i])
	}

	return power
}

// SpectrumResult 频谱分析结果
type SpectrumResult struct {
	MaxFrequency    float64   // 最高有效频率
	CutoffFrequency float64   // 截断频率
	IsFake          bool      // 是否为假无损
	Details         string    // 详细说明
	PowerSpectrum   []float64 // 功率谱（用于进一步分析）
}

// analyzeFrequencyContent 分析频率内容
func (s *SpectrumAnalyzer) analyzeFrequencyContent(powerSpectrum []float64) *SpectrumResult {
	// 频率分辨率
	freqResolution := float64(s.sampleRate) / float64(len(powerSpectrum)*2)

	// 找到最高有效频率
	maxFreq := s.findMaxEffectiveFrequency(powerSpectrum, freqResolution)

	// 检测是否存在明显的频率截断
	cutoffFreq := s.detectFrequencyCutoff(powerSpectrum, freqResolution)

	// 判断是否为假无损
	isFake, details := s.determineFakeStatus(maxFreq, cutoffFreq)

	return &SpectrumResult{
		MaxFrequency:    maxFreq,
		CutoffFrequency: cutoffFreq,
		IsFake:          isFake,
		Details:         details,
		PowerSpectrum:   powerSpectrum,
	}
}

// findMaxEffectiveFrequency 找到最高有效频率
func (s *SpectrumAnalyzer) findMaxEffectiveFrequency(powerSpectrum []float64, freqResolution float64) float64 {
	// 计算噪声基底
	noiseFloor := s.calculateNoiseFloor(powerSpectrum)

	// 从高频往低频搜索，找到最后一个显著高于噪声基底的频率
	threshold := noiseFloor * 10 // 阈值设为噪声基底的10倍

	for i := len(powerSpectrum) - 1; i >= 0; i-- {
		if powerSpectrum[i] > threshold {
			return float64(i) * freqResolution
		}
	}

	return 0
}

// detectFrequencyCutoff 检测频率截断
func (s *SpectrumAnalyzer) detectFrequencyCutoff(powerSpectrum []float64, freqResolution float64) float64 {
	// 寻找功率急剧下降的点
	maxPower := 0.0
	for _, power := range powerSpectrum {
		if power > maxPower {
			maxPower = power
		}
	}

	// 阈值设为最大功率的1%
	threshold := maxPower * 0.01

	// 从高频段开始寻找连续低于阈值的点
	consecutiveLow := 0
	requiredConsecutive := 10 // 需要连续10个点都低于阈值

	for i := len(powerSpectrum) - 1; i >= 0; i-- {
		if powerSpectrum[i] < threshold {
			consecutiveLow++
			if consecutiveLow >= requiredConsecutive {
				return float64(i+requiredConsecutive) * freqResolution
			}
		} else {
			consecutiveLow = 0
		}
	}

	return float64(len(powerSpectrum)) * freqResolution
}

// calculateNoiseFloor 计算噪声基底
func (s *SpectrumAnalyzer) calculateNoiseFloor(powerSpectrum []float64) float64 {
	// 取功率谱的最后10%作为噪声基底的估计
	startIdx := len(powerSpectrum) * 9 / 10

	sum := 0.0
	count := 0

	for i := startIdx; i < len(powerSpectrum); i++ {
		sum += powerSpectrum[i]
		count++
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

// determineFakeStatus 判断是否为假无损
func (s *SpectrumAnalyzer) determineFakeStatus(maxFreq, cutoffFreq float64) (bool, string) {
	// 常见的有损编码截断频率
	commonCutoffs := map[float64]string{
		16000: "MP3 128kbps",
		17000: "MP3 160kbps",
		19000: "MP3 192kbps",
		20000: "MP3 256kbps",
		21000: "MP3 320kbps",
	}

	// 检查是否接近已知的有损编码截断频率
	for cutoff, format := range commonCutoffs {
		if math.Abs(maxFreq-cutoff) < 500 { // 500Hz的容差
			return true, fmt.Sprintf("检测到%s格式的典型截断频率 (%.0f Hz)", format, maxFreq)
		}
	}

	// 如果最高频率低于18kHz，很可能是假无损
	if maxFreq < 18000 {
		return true, fmt.Sprintf("最高有效频率过低 (%.0f Hz)，可能从有损格式转换而来", maxFreq)
	}

	// 如果存在明显的频率截断
	if cutoffFreq < float64(s.sampleRate)/2*0.9 { // 截断频率低于奈奎斯特频率的90%
		return true, fmt.Sprintf("在 %.0f Hz 附近检测到明显的频率截断", cutoffFreq)
	}

	return false, fmt.Sprintf("频谱正常，最高有效频率 %.0f Hz", maxFreq)
}

// nearestPowerOf2 找到最接近的2的幂
func nearestPowerOf2(n int) int {
	power := 1
	for power < n {
		power <<= 1
	}
	return power
}
