# Audio Loss Checker 使用示例

## 基本使用

### 分析单个文件
```bash
# 分析单个音频文件
.\audio-loss-checker.exe path\to\audio\file.flac

# 使用自定义截断频率阈值
.\audio-loss-checker.exe --cutoff 19000 path\to\file.flac
```

### 分析整个目录
```bash
# 分析整个音乐目录
.\audio-loss-checker.exe C:\Music

# 使用并发处理加速分析
.\audio-loss-checker.exe -j 8 C:\Music
```

## 输出控制

### 静默模式
```bash
# 只输出假无损文件的路径
.\audio-loss-checker.exe --quiet C:\Music

# 将结果保存到文件
.\audio-loss-checker.exe --quiet C:\Music > fake_files.txt
```

### 只显示可疑文件
```bash
# 只显示假无损文件的详细信息
.\audio-loss-checker.exe --only-fake C:\Music
```

### JSON 输出
```bash
# 以JSON格式输出结果
.\audio-loss-checker.exe --json C:\Music

# 保存JSON结果到文件
.\audio-loss-checker.exe --json C:\Music > analysis_results.json
```

## 实际使用场景

### 场景1: 快速检查可疑文件
```bash
# 扫描音乐库，只关心有问题的文件，使用4个线程加速
.\audio-loss-checker.exe --only-fake -j 4 "D:\Music Library" > suspicious_files.log
```

### 场景2: 完整分析报告
```bash
# 生成完整的JSON分析报告
.\audio-loss-checker.exe --json -j 8 "C:\Users\Music" > full_analysis.json
```

### 场景3: 批量检查特定格式
```bash
# 先找出所有FLAC文件，然后分析
# (需要配合其他工具或脚本)
```

## 输出说明

### 正常输出示例
```
=== example.flac ===
路径: C:\Music\example.flac
格式: FLAC
状态: OK
采样率: 44100 Hz
位深度: 16 bit
声道数: 2
时长: 245.67 秒
标题: Example Song
艺术家: Example Artist
专辑: Example Album
最高有效频率: 20950 Hz
分析结果: 频谱正常，最高有效频率 20950 Hz
✅ 文件看起来是真实的无损音频
```

### 假无损检测示例
```
=== fake.flac ===
路径: C:\Music\fake.flac
格式: FLAC
状态: FAKE
采样率: 44100 Hz
位深度: 16 bit
声道数: 2
时长: 180.23 秒
最高有效频率: 16000 Hz
截断频率: 16000 Hz
分析结果: 检测到MP3 128kbps格式的典型截断频率 (16000 Hz)
⚠️  警告: 这可能是一个假无损文件！
```

## 技术原理

本工具通过以下方法检测假无损音频：

1. **频谱分析**: 使用FFT分析音频频谱
2. **截断检测**: 查找高频部分的异常截断
3. **模式识别**: 识别常见有损编码的截断模式
4. **阈值判断**: 根据设定阈值判断文件真伪

### 常见截断频率
- **16 kHz**: MP3 128kbps
- **17 kHz**: MP3 160kbps  
- **19 kHz**: MP3 192kbps
- **20 kHz**: MP3 256kbps
- **21 kHz**: MP3 320kbps

## 注意事项

1. **分析准确性**: 工具基于频谱分析，可能存在误判
2. **文件格式**: 目前支持WAV和FLAC，ALAC和APE格式正在开发中
3. **处理时间**: 大文件分析需要时间，建议使用并发选项
4. **结果解读**: 建议结合听感和其他工具综合判断

## 故障排除

### 常见错误
- **"不支持的音频格式"**: 检查文件扩展名和格式
- **"解码失败"**: 文件可能损坏或格式不正确
- **"频谱分析失败"**: 音频数据可能有问题

### 性能优化
- 使用`-j`参数增加并发数（建议不超过CPU核心数的2倍）
- 对于大型音乐库，建议分批处理
- 使用SSD存储可以提高I/O性能