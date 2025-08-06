package decoder

import (
	"fmt"
	"os"
	"time"

	"audio-loss-checker/internal/types"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// WAVDecoder WAV格式解码器
type WAVDecoder struct{}

// WAVFile WAV文件实现
type WAVFile struct {
	decoder    *wav.Decoder
	file       *os.File
	format     *audio.Format
	sampleRate int
	bitDepth   int
	channels   int
	duration   time.Duration
	samples    []float64
}

// SupportedFormats 返回支持的格式
func (d *WAVDecoder) SupportedFormats() []string {
	return []string{"wav"}
}

// Decode 解码WAV文件
func (d *WAVDecoder) Decode(filePath string) (types.AudioFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开WAV文件失败: %w", err)
	}

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		file.Close()
		return nil, fmt.Errorf("无效的WAV文件: %s", filePath)
	}

	// 获取音频格式信息
	format := &audio.Format{
		NumChannels: int(decoder.NumChans),
		SampleRate:  int(decoder.SampleRate),
	}
	bitDepth := int(decoder.BitDepth)

	// 计算时长
	duration := time.Duration(float64(decoder.PCMLen()) / float64(format.SampleRate) * float64(time.Second))

	wavFile := &WAVFile{
		decoder:    decoder,
		file:       file,
		format:     format,
		sampleRate: format.SampleRate,
		bitDepth:   bitDepth,
		channels:   format.NumChannels,
		duration:   duration,
	}

	return wavFile, nil
}

// GetFormat 获取格式名称
func (w *WAVFile) GetFormat() string {
	return "WAV"
}

// GetSampleRate 获取采样率
func (w *WAVFile) GetSampleRate() int {
	return w.sampleRate
}

// GetBitDepth 获取位深度
func (w *WAVFile) GetBitDepth() int {
	return w.bitDepth
}

// GetChannels 获取声道数
func (w *WAVFile) GetChannels() int {
	return w.channels
}

// GetDuration 获取时长
func (w *WAVFile) GetDuration() time.Duration {
	return w.duration
}

// GetSamples 获取音频采样数据
func (w *WAVFile) GetSamples() ([]float64, error) {
	if w.samples != nil {
		return w.samples, nil
	}

	// 读取所有音频数据
	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: w.channels,
			SampleRate:  w.sampleRate,
		},
	}

	for {
		n, err := w.decoder.PCMBuffer(buf)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
	}

	// 转换为float64格式
	samples := make([]float64, len(buf.Data))
	maxVal := float64(int(1) << uint(w.bitDepth-1))

	for i, sample := range buf.Data {
		samples[i] = float64(sample) / maxVal
	}

	w.samples = samples
	return samples, nil
}

// GetMetadata 获取元数据
func (w *WAVFile) GetMetadata() types.AudioMetadata {
	// WAV文件的元数据支持有限，这里返回基本信息
	return types.AudioMetadata{
		Duration: w.duration.String(),
	}
}

// Close 关闭文件
func (w *WAVFile) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
