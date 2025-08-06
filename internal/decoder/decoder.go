package decoder

import (
	"fmt"
	"path/filepath"
	"strings"

	"audio-loss-checker/internal/types"
)

// AudioDecoder 音频解码器接口
type AudioDecoder interface {
	Decode(filePath string) (types.AudioFile, error)
	SupportedFormats() []string
}

// DecoderRegistry 解码器注册表
type DecoderRegistry struct {
	decoders map[string]AudioDecoder
}

// NewDecoderRegistry 创建新的解码器注册表
func NewDecoderRegistry() *DecoderRegistry {
	registry := &DecoderRegistry{
		decoders: make(map[string]AudioDecoder),
	}

	// 注册支持的解码器
	registry.Register(&WAVDecoder{})
	registry.Register(&FLACDecoder{})

	// TODO: 添加 ALAC 和 APE 解码器
	// ALAC: 可考虑使用 FFmpeg 绑定或 github.com/go-audio/m4a
	// APE: 可考虑使用 FFmpeg 绑定或寻找专用的 APE 解码库
	// registry.Register(&ALACDecoder{})
	// registry.Register(&APEDecoder{})

	return registry
}

// Register 注册解码器
func (r *DecoderRegistry) Register(decoder AudioDecoder) {
	for _, format := range decoder.SupportedFormats() {
		r.decoders[strings.ToLower(format)] = decoder
	}
}

// GetDecoder 根据文件扩展名获取解码器
func (r *DecoderRegistry) GetDecoder(filePath string) (AudioDecoder, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return nil, fmt.Errorf("无法确定文件格式: %s", filePath)
	}

	// 移除点号
	ext = ext[1:]

	decoder, exists := r.decoders[ext]
	if !exists {
		return nil, fmt.Errorf("不支持的音频格式: %s", ext)
	}

	return decoder, nil
}

// DecodeFile 解码音频文件
func (r *DecoderRegistry) DecodeFile(filePath string) (types.AudioFile, error) {
	decoder, err := r.GetDecoder(filePath)
	if err != nil {
		return nil, err
	}

	return decoder.Decode(filePath)
}
