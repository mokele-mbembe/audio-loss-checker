package decoder

import (
	"fmt"
	"os"
	"time"

	"audio-loss-checker/internal/types"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/meta"
)

// FLACDecoder FLAC格式解码器
type FLACDecoder struct{}

// FLACFile FLAC文件实现
type FLACFile struct {
	stream     *flac.Stream
	file       *os.File
	sampleRate int
	bitDepth   int
	channels   int
	duration   time.Duration
	samples    []float64
	metadata   types.AudioMetadata
}

// SupportedFormats 返回支持的格式
func (d *FLACDecoder) SupportedFormats() []string {
	return []string{"flac"}
}

// Decode 解码FLAC文件
func (d *FLACDecoder) Decode(filePath string) (types.AudioFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开FLAC文件失败: %w", err)
	}

	stream, err := flac.New(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("解析FLAC文件失败: %w", err)
	}

	info := stream.Info
	if info == nil {
		file.Close()
		return nil, fmt.Errorf("无法读取FLAC信息: %s", filePath)
	}

	// 计算时长
	duration := time.Duration(float64(info.NSamples) / float64(info.SampleRate) * float64(time.Second))

	flacFile := &FLACFile{
		stream:     stream,
		file:       file,
		sampleRate: int(info.SampleRate),
		bitDepth:   int(info.BitsPerSample),
		channels:   int(info.NChannels),
		duration:   duration,
	}

	// 解析元数据
	flacFile.parseMetadata()

	return flacFile, nil
}

// parseMetadata 解析FLAC元数据
func (f *FLACFile) parseMetadata() {
	for _, block := range f.stream.Blocks {
		if block.Header.Type == meta.TypeVorbisComment {
			if comment, ok := block.Body.(*meta.VorbisComment); ok {
				f.metadata = types.AudioMetadata{
					Title:    getVorbisTag(comment, "TITLE"),
					Artist:   getVorbisTag(comment, "ARTIST"),
					Album:    getVorbisTag(comment, "ALBUM"),
					Year:     getVorbisTag(comment, "DATE"),
					Genre:    getVorbisTag(comment, "GENRE"),
					Duration: f.duration.String(),
				}
			}
		}
	}
}

// getVorbisTag 获取Vorbis注释标签
func getVorbisTag(comment *meta.VorbisComment, tag string) string {
	for _, field := range comment.Tags {
		if field[0] == tag {
			return field[1]
		}
	}
	return ""
}

// GetFormat 获取格式名称
func (f *FLACFile) GetFormat() string {
	return "FLAC"
}

// GetSampleRate 获取采样率
func (f *FLACFile) GetSampleRate() int {
	return f.sampleRate
}

// GetBitDepth 获取位深度
func (f *FLACFile) GetBitDepth() int {
	return f.bitDepth
}

// GetChannels 获取声道数
func (f *FLACFile) GetChannels() int {
	return f.channels
}

// GetDuration 获取时长
func (f *FLACFile) GetDuration() time.Duration {
	return f.duration
}

// GetSamples 获取音频采样数据
func (f *FLACFile) GetSamples() ([]float64, error) {
	if f.samples != nil {
		return f.samples, nil
	}

	var allSamples []float64
	maxVal := float64(int(1) << uint(f.bitDepth-1))

	// 读取所有音频帧
	for {
		frame, err := f.stream.ParseNext()
		if err != nil {
			break
		}

		// 将所有声道的数据合并（简化处理，实际应用中可能需要更复杂的处理）
		for i := 0; i < len(frame.Subframes[0].Samples); i++ {
			for ch := 0; ch < f.channels; ch++ {
				sample := float64(frame.Subframes[ch].Samples[i]) / maxVal
				allSamples = append(allSamples, sample)
			}
		}
	}

	f.samples = allSamples
	return allSamples, nil
}

// GetMetadata 获取元数据
func (f *FLACFile) GetMetadata() types.AudioMetadata {
	return f.metadata
}

// Close 关闭文件
func (f *FLACFile) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
