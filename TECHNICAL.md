# Audio Loss Checker 技术原理

本文档详细介绍 Audio Loss Checker 的技术实现原理和算法。

## 目录
- [概述](#概述)
- [核心算法](#核心算法)
- [音频解码](#音频解码)
- [频谱分析](#频谱分析)
- [假无损检测](#假无损检测)
- [性能优化](#性能优化)
- [技术栈](#技术栈)

## 概述

Audio Loss Checker 通过分析音频文件的频谱特征来检测"假无损"文件。所谓"假无损"，是指从有损音频格式（如MP3、AAC等）转换而来的无损格式文件，这些文件虽然具有无损格式的外壳，但实际音质已经受损。

### 检测原理

有损音频编码器为了压缩文件大小，会移除人耳不易察觉的音频信息，主要表现为：

1. **高频截断**: 移除高频部分（如16kHz以上）
2. **心理声学模型**: 基于掩蔽效应移除被"掩盖"的频率
3. **量化噪声**: 引入轻微的量化误差

这些特征在频域中表现为明显的模式，可以通过频谱分析检测出来。

## 核心算法

### 1. 音频解码流程

```
音频文件 → 格式检测 → 解码器选择 → PCM数据提取 → 频谱分析
```

### 2. 频谱分析流程

```
PCM采样 → 窗函数处理 → FFT变换 → 功率谱计算 → 特征提取 → 模式识别
```

## 音频解码

### WAV格式解码

WAV是未压缩的PCM音频格式，解码相对简单：

```go
// 核心解码逻辑
decoder := wav.NewDecoder(file)
format := &audio.Format{
    NumChannels: int(decoder.NumChans),
    SampleRate:  int(decoder.SampleRate),
}
```

**关键步骤：**
1. 读取WAV文件头，获取采样率、位深度、声道数
2. 提取PCM数据
3. 转换为归一化的float64数组（-1.0 到 1.0）

### FLAC格式解码

FLAC是无损压缩格式，需要解压缩：

```go
// 使用github.com/mewkiz/flac库
stream, err := flac.New(file)
info := stream.Info

// 逐帧解码
for {
    frame, err := stream.ParseNext()
    // 处理音频帧数据
}
```

**关键特性：**
- 支持元数据提取（标题、艺术家等）
- 可变位深度支持（16/24位）
- 逐帧解码，内存效率高

## 频谱分析

### 1. 预处理

#### 采样窗口选择
```go
// 选择分析窗口（避免静音段）
startIdx := len(samples) / 4          // 跳过开头25%
endIdx := startIdx + windowSize       // 取固定窗口大小
window := samples[startIdx:endIdx]
```

#### 汉明窗函数
```go
// 应用汉明窗减少频谱泄漏
func applyHammingWindow(samples []float64) []float64 {
    windowed := make([]float64, len(samples))
    n := len(samples)
    
    for i, sample := range samples {
        // 汉明窗函数: w(n) = 0.54 - 0.46 * cos(2π * n / (N-1))
        window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
        windowed[i] = sample * window
    }
    return windowed
}
```

**汉明窗的作用：**
- 减少频谱泄漏（spectral leakage）
- 改善频率分辨率
- 降低旁瓣效应

### 2. FFT变换

```go
// 使用github.com/mjibson/go-dsp/fft
spectrum := fft.FFTReal(windowedSamples)
```

**FFT参数：**
- **窗口大小**: 8192 samples（提供良好的频率分辨率）
- **频率分辨率**: sampleRate / windowSize
- **奈奎斯特频率**: sampleRate / 2

### 3. 功率谱计算

```go
// 计算功率谱密度
func calculatePowerSpectrum(spectrum []complex128) []float64 {
    power := make([]float64, len(spectrum)/2) // 只需要一半（对称性）
    
    for i := 0; i < len(power); i++ {
        // 功率 = |复数|²
        power[i] = cmplx.Abs(spectrum[i]) * cmplx.Abs(spectrum[i])
    }
    return power
}
```

## 假无损检测

### 1. 最高有效频率检测

```go
func findMaxEffectiveFrequency(powerSpectrum []float64, freqResolution float64) float64 {
    // 计算噪声基底
    noiseFloor := calculateNoiseFloor(powerSpectrum)
    
    // 阈值设为噪声基底的10倍
    threshold := noiseFloor * 10
    
    // 从高频往低频搜索
    for i := len(powerSpectrum) - 1; i >= 0; i-- {
        if powerSpectrum[i] > threshold {
            return float64(i) * freqResolution
        }
    }
    return 0
}
```

### 2. 频率截断检测

```go
func detectFrequencyCutoff(powerSpectrum []float64, freqResolution float64) float64 {
    // 寻找功率急剧下降的点
    maxPower := findMaxPower(powerSpectrum)
    threshold := maxPower * 0.01  // 1%阈值
    
    // 寻找连续低于阈值的区域
    consecutiveLow := 0
    requiredConsecutive := 10  // 需要连续10个点
    
    for i := len(powerSpectrum) - 1; i >= 0; i-- {
        if powerSpectrum[i] < threshold {
            consecutiveLow++
            if consecutiveLow >= requiredConsecutive {
                return float64(i + requiredConsecutive) * freqResolution
            }
        } else {
            consecutiveLow = 0
        }
    }
    return float64(len(powerSpectrum)) * freqResolution
}
```

### 3. 模式识别

#### 已知有损格式的截断模式：

| 格式 | 典型截断频率 | 特征 |
|------|-------------|------|
| MP3 128kbps | ~16.0 kHz | 硬截断 |
| MP3 160kbps | ~17.0 kHz | 硬截断 |
| MP3 192kbps | ~19.0 kHz | 硬截断 |
| MP3 256kbps | ~20.0 kHz | 硬截断 |
| MP3 320kbps | ~21.0 kHz | 软截断 |
| AAC 128kbps | ~15.5 kHz | 渐进截断 |

```go
func determineFakeStatus(maxFreq, cutoffFreq float64) (bool, string) {
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
        if math.Abs(maxFreq-cutoff) < 500 { // 500Hz容差
            return true, fmt.Sprintf("检测到%s格式的典型截断频率", format)
        }
    }
    
    // 通用低频判断
    if maxFreq < 18000 {
        return true, fmt.Sprintf("最高有效频率过低 (%.0f Hz)", maxFreq)
    }
    
    return false, fmt.Sprintf("频谱正常，最高有效频率 %.0f Hz", maxFreq)
}
```

## 性能优化

### 1. 并发处理

```go
// 工作池模式
jobs := make(chan string, len(filePaths))
results := make(chan *types.AnalysisResult, len(filePaths))

// 启动多个工作协程
var wg sync.WaitGroup
for i := 0; i < concurrency; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for filePath := range jobs {
            result := analyzeFile(filePath)
            results <- result
        }
    }()
}
```

**优化策略：**
- 使用工作池模式避免频繁创建goroutine
- 通道缓冲区大小与文件数量匹配
- 合理设置并发数（通常为CPU核心数的1-2倍）

### 2. 内存管理

```go
// 流式处理，避免一次性加载整个文件
func (f *FLACFile) GetSamples() ([]float64, error) {
    if f.samples != nil {
        return f.samples, nil  // 缓存结果
    }
    
    var allSamples []float64
    // 逐帧处理，控制内存使用
    for {
        frame, err := f.stream.ParseNext()
        if err != nil {
            break
        }
        // 处理当前帧...
    }
    
    f.samples = allSamples  // 缓存结果
    return allSamples, nil
}
```

### 3. FFT优化

- **窗口大小选择**: 8192样本平衡了频率分辨率和计算效率
- **实数FFT**: 使用`FFTReal`而非复数FFT，减少一半计算量
- **预分配内存**: 避免频繁的内存分配

## 技术栈

### 核心依赖

```go
require (
    github.com/spf13/cobra v1.8.0                    // CLI框架
    github.com/mewkiz/flac v1.0.7                    // FLAC解码
    github.com/go-audio/wav v1.1.0                   // WAV解码
    github.com/go-audio/audio v1.0.0                 // 音频基础类型
    github.com/mjibson/go-dsp v0.0.0-20180508042940  // 数字信号处理
    github.com/schollz/progressbar/v3 v3.14.1        // 进度条显示
)
```

### 架构设计

```
cmd/                 # CLI命令层
├── root.go         # 主命令和参数解析

internal/           # 内部实现
├── types/          # 类型定义
│   └── types.go    # 数据结构
├── decoder/        # 音频解码层
│   ├── decoder.go  # 解码器注册表
│   ├── wav.go      # WAV解码器
│   └── flac.go     # FLAC解码器
└── analyzer/       # 分析层
    ├── analyzer.go # 主分析器
    └── spectrum.go # 频谱分析器
```

## 算法限制与改进方向

### 当前限制

1. **单一特征检测**: 主要依赖频率截断，可能误判某些特殊录音
2. **静态阈值**: 使用固定的频率阈值，不够灵活
3. **窗口选择**: 固定选择中间部分，可能错过关键信息

### 改进方向

1. **多特征融合**:
   - 频谱平坦度分析
   - 相位相关性检测
   - 时域统计特征

2. **机器学习增强**:
   - 训练分类模型
   - 特征自动提取
   - 自适应阈值

3. **更精细的分析**:
   - 多窗口分析
   - 时频联合分析
   - 心理声学模型

## 参考资料

- [Digital Signal Processing - Oppenheim & Schafer](https://www.amazon.com/Digital-Signal-Processing-Alan-Oppenheim/dp/0131988425)
- [FLAC Format Specification](https://xiph.org/flac/format.html)
- [WAV File Format](http://soundfile.sapp.org/doc/WaveFormat/)
- [FFT Algorithm](https://en.wikipedia.org/wiki/Fast_Fourier_transform)
- [Window Functions](https://en.wikipedia.org/wiki/Window_function)

---

*本文档持续更新，欢迎提出改进建议。*