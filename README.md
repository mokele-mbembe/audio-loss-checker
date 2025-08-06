# Audio Loss Checker

一个简单的CLI工具，用于检测无损音频文件是否真的是无损格式。

## 支持格式
- **WAV** - Windows音频文件 ✅
- **FLAC** - 自由无损音频编解码器 ✅
- **ALAC** - Apple无损音频编解码器 🚧 (开发中)
- **APE** - Monkey's Audio 🚧 (开发中)

## 命令行参数

### 1. 输出控制 (Output Control)

#### `-q, --quiet`
静默模式，仅输出被判定为"假无损"的文件路径，每行一个。

```bash
# 扫描整个目录，只列出有问题的文件路径
./audio-loss-checker --quiet /mnt/music

# 示例输出:
# /mnt/music/album1/track01.flac
# /mnt/music/album3/track05.flac
```

#### `--only-fake`
只显示被判定为"假无损"文件的完整分析报告。

```bash
# 扫描目录，并详细显示每个可疑文件的信息
./audio-loss-checker --only-fake /mnt/music
```

#### `--json`
以JSON格式输出所有文件的分析结果。

```bash
# 以JSON格式扫描目录，并将结果重定向到文件
./audio-loss-checker --json /mnt/music > analysis_result.json
```

**JSON输出示例:**
```json
{
  "filePath": "/mnt/music/real.flac",
  "format": "FLAC",
  "metadata": { "title": "Real Song", "artist": "Good Artist" },
  "status": "OK",
  "analysis": { "isFake": false, "details": "频谱正常，可能是真实的无损音乐" }
}
{
  "filePath": "/mnt/music/fake.flac",
  "format": "FLAC",
  "metadata": { "title": "Fake Song", "artist": "Bad Converter" },
  "status": "FAKE",
  "analysis": { "isFake": true, "cutoffHz": 16054, "details": "在 16054 Hz 附近有明显截断" }
}
```

### 2. 分析调整 (Analysis Tuning)

#### `--cutoff <frequency>`
设置自定义的频率截断阈值（单位Hz）。如果检测到的频谱最高有效频率低于此值，则判定为假无损。

```bash
# 使用更严格的19kHz作为判断标准
./audio-loss-checker --cutoff 19000 /path/to/file.flac

# 使用更宽松的17kHz标准，用于排查一些低质量录音
./audio-loss-checker --cutoff 17000 /path/to/file.flac
```

### 3. 性能与通用选项 (Performance & General)

#### `-j <number>, --concurrency <number>`
设置并发处理文件的数量，可以显著加快扫描大型目录的速度。

```bash
# 使用8个并发任务扫描目录
./audio-loss-checker -j 8 /mnt/huge_music_library
```

#### `-v, --version`
显示程序版本。

```bash
./audio-loss-checker --version
# 输出: audio-loss-checker version 1.1.0
```

#### `-h, --help`
显示帮助菜单，列出所有可用命令和选项。

```bash
./audio-loss-checker --help
```

## 使用示例

### 组合参数使用
**场景:** 快速扫描大型音乐库，查找可疑的"假无损"文件

假设你想要：
- 使用4个CPU核心进行并发扫描以节省时间
- 只关心那些有问题的"假无损"文件
- 将完整报告保存到日志文件中供日后查看

```bash
./audio-loss-checker --only-fake -j 4 /mnt/nas/music > suspicious_files_report.txt
```

这个命令会高效地完成扫描，并将所有可疑文件的详细分析保存在 `suspicious_files_report.txt` 文件中。

## 安装与编译

### 预编译版本
从 [Releases](../../releases) 页面下载适合你操作系统的预编译版本。

### 从源码编译

#### 前置要求
- Go 1.24+ 

#### 编译步骤
```bash
# 克隆仓库
git clone https://github.com/your-username/audio-loss-checker.git
cd audio-loss-checker

# 使用 Makefile 编译（推荐）
make help          # 查看所有可用命令
make build         # 编译当前平台版本
make build-all     # 编译所有平台版本
make install       # 安装到系统

# 或者直接使用 Go 命令
go mod tidy
go build -o audio-loss-checker
```

#### Makefile 命令说明
```bash
make build          # 编译当前平台版本
make build-all      # 编译所有平台版本 (Windows/Linux/macOS)
make build-windows  # 仅编译 Windows 版本
make build-linux    # 仅编译 Linux 版本  
make build-darwin   # 仅编译 macOS 版本
make clean          # 清理构建文件
make test           # 运行测试
make fmt            # 格式化代码
make install        # 安装到系统
make release        # 创建发布包
```

## 技术原理

### 检测方法
1. **FFT频谱分析**: 对音频进行快速傅里叶变换
2. **高频截断检测**: 识别人工截断的频率边界
3. **模式匹配**: 对比已知有损编码的频谱特征
4. **阈值判断**: 基于用户设定或默认阈值进行判断

> 📖 详细技术原理请参考 [TECHNICAL.md](TECHNICAL.md)

### 支持格式
- ✅ **WAV**: 完全支持
- ✅ **FLAC**: 完全支持，包括元数据解析
- 🚧 **ALAC**: 基础支持（开发中）
- 🚧 **APE**: 基础支持（开发中）

### 检测准确性
- **高准确性**: MP3转换的FLAC文件（典型截断模式）
- **中等准确性**: 其他有损格式转换的文件
- **注意**: 某些高质量录音可能被误判，建议结合听感判断

## 开发

### 项目结构
```
audio-loss-checker/
├── cmd/                    # CLI命令定义
├── internal/
│   ├── analyzer/          # 音频分析器
│   ├── decoder/           # 音频解码器
│   └── types/             # 类型定义
├── examples.md            # 使用示例
├── TECHNICAL.md           # 技术原理文档
├── Makefile              # 跨平台构建脚本
└── README.md
```

### 贡献
欢迎提交Issue和Pull Request！

## 许可证
本项目采用 [LICENSE](LICENSE) 许可证。
