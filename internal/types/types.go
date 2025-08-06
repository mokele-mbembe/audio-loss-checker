package types

import "time"

// AnalyzerConfig 分析器配置
type AnalyzerConfig struct {
	CutoffFreq  float64 // 频率截断阈值 (Hz)
	Concurrency int     // 并发数
	Quiet       bool    // 静默模式
	OnlyFake    bool    // 只显示假无损
	JSONOutput  bool    // JSON输出格式
}

// AudioMetadata 音频元数据
type AudioMetadata struct {
	Title    string `json:"title,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	Year     string `json:"year,omitempty"`
	Genre    string `json:"genre,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// AnalysisDetails 详细分析结果
type AnalysisDetails struct {
	IsFake       bool    `json:"isFake"`
	CutoffHz     float64 `json:"cutoffHz,omitempty"`
	Details      string  `json:"details"`
	SampleRate   int     `json:"sampleRate"`
	BitDepth     int     `json:"bitDepth"`
	Channels     int     `json:"channels"`
	Duration     float64 `json:"duration"`
	MaxFrequency float64 `json:"maxFrequency"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	FilePath string          `json:"filePath"`
	Format   string          `json:"format"`
	Metadata AudioMetadata   `json:"metadata"`
	Status   string          `json:"status"` // "OK", "FAKE", "ERROR"
	Analysis AnalysisDetails `json:"analysis"`
	Error    string          `json:"error,omitempty"`
}

// AudioFile 音频文件接口
type AudioFile interface {
	GetFormat() string
	GetSampleRate() int
	GetBitDepth() int
	GetChannels() int
	GetDuration() time.Duration
	GetSamples() ([]float64, error)
	GetMetadata() AudioMetadata
	Close() error
}
