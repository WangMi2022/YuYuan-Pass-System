package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	stdDraw "image/draw"
	"image/jpeg"
	_ "image/png"
	"math"

	xDraw "golang.org/x/image/draw"
)

const (
	maxMultimodalInitialImageSize = 640 << 10
	maxMultimodalRetryImageSize   = 384 << 10
	minMultimodalRetryImageSize   = 64 << 10
	maxMultimodalDecodedPixels    = 24_000_000
	maxMultimodalImageEdge        = 2400
	minMultimodalJPEGQuality      = 68
	maxMultimodalJPEGQuality      = 90
)

var multimodalImageProcessingSlots = make(chan struct{}, 2)

func prepareMultimodalImage(ctx context.Context, input Input, maxBytes int) (Input, error) {
	if maxBytes <= 0 {
		return Input{}, errors.New("多模态图片大小限制不正确")
	}
	if len(input.Data) <= maxBytes {
		return input, nil
	}
	return transcodeMultimodalImage(ctx, input, maxBytes)
}

func prepareMultimodalRetryImage(ctx context.Context, input Input) (Input, error) {
	maxBytes := maxMultimodalRetryImageSize
	if reducedSize := len(input.Data) * 3 / 4; reducedSize < maxBytes {
		maxBytes = max(minMultimodalRetryImageSize, reducedSize)
	}
	if len(input.Data) <= maxBytes {
		return input, nil
	}
	return transcodeMultimodalImage(ctx, input, maxBytes)
}

func transcodeMultimodalImage(ctx context.Context, input Input, maxBytes int) (Input, error) {
	select {
	case multimodalImageProcessingSlots <- struct{}{}:
		defer func() { <-multimodalImageProcessingSlots }()
	case <-ctx.Done():
		return Input{}, ctx.Err()
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(input.Data))
	if err != nil {
		return Input{}, fmt.Errorf("发票图片无法读取: %w", err)
	}
	if format != "jpeg" && format != "png" {
		return Input{}, errors.New("多模态模型仅支持 JPG 或 PNG 图片")
	}
	if config.Width <= 0 || config.Height <= 0 ||
		int64(config.Width)*int64(config.Height) > maxMultimodalDecodedPixels {
		return Input{}, errors.New("发票图片分辨率过高，请压缩到 2400 万像素以内后重试")
	}

	source, _, err := image.Decode(bytes.NewReader(input.Data))
	if err != nil {
		return Input{}, fmt.Errorf("发票图片无法读取: %w", err)
	}
	if format == "jpeg" {
		source = applyEXIFOrientation(source, jpegEXIFOrientation(input.Data))
	}
	prepared, err := encodeInvoiceJPEGWithinBudget(flattenAndResizeOnWhite(source, maxMultimodalImageEdge), maxBytes)
	if err != nil {
		return Input{}, err
	}
	input.ContentType = "image/jpeg"
	input.Data = prepared
	return input, nil
}

func flattenAndResizeOnWhite(source image.Image, maxEdge int) *image.RGBA {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if longest := max(width, height); longest > maxEdge {
		ratio := float64(maxEdge) / float64(longest)
		width = max(1, int(math.Round(float64(width)*ratio)))
		height = max(1, int(math.Round(float64(height)*ratio)))
	}
	flattened := image.NewRGBA(image.Rect(0, 0, width, height))
	stdDraw.Draw(flattened, flattened.Bounds(), image.NewUniform(color.White), image.Point{}, stdDraw.Src)
	if width == bounds.Dx() && height == bounds.Dy() {
		stdDraw.Draw(flattened, flattened.Bounds(), source, bounds.Min, stdDraw.Over)
	} else {
		xDraw.CatmullRom.Scale(flattened, flattened.Bounds(), source, bounds, xDraw.Over, nil)
	}
	return flattened
}

func encodeInvoiceJPEGWithinBudget(source *image.RGBA, maxBytes int) ([]byte, error) {
	current := source
	for attempt := 0; attempt < 8; attempt++ {
		encoded, smallestSize, ok, err := highestQualityJPEGWithinBudget(current, maxBytes)
		if err != nil {
			return nil, fmt.Errorf("发票图片压缩失败: %w", err)
		}
		if ok {
			return encoded, nil
		}

		ratio := math.Sqrt(float64(maxBytes)/float64(smallestSize)) * 0.94
		ratio = math.Max(0.5, math.Min(0.88, ratio))
		width := max(1, int(math.Round(float64(current.Bounds().Dx())*ratio)))
		height := max(1, int(math.Round(float64(current.Bounds().Dy())*ratio)))
		if width == current.Bounds().Dx() && height == current.Bounds().Dy() {
			break
		}
		current = resizeImage(current, width, height)
	}
	return nil, errors.New("发票图片压缩后仍然过大，请降低图片分辨率后重试")
}

func highestQualityJPEGWithinBudget(source image.Image, maxBytes int) ([]byte, int, bool, error) {
	lowest, err := encodeJPEG(source, minMultimodalJPEGQuality)
	if err != nil {
		return nil, 0, false, err
	}
	if len(lowest) > maxBytes {
		return nil, len(lowest), false, nil
	}

	best := lowest
	low, high := minMultimodalJPEGQuality+1, maxMultimodalJPEGQuality
	for low <= high {
		quality := low + (high-low)/2
		candidate, encodeErr := encodeJPEG(source, quality)
		if encodeErr != nil {
			return nil, 0, false, encodeErr
		}
		if len(candidate) <= maxBytes {
			best = candidate
			low = quality + 1
		} else {
			high = quality - 1
		}
	}
	return best, len(lowest), true, nil
}

func encodeJPEG(source image.Image, quality int) ([]byte, error) {
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, source, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}

func resizeImage(source image.Image, width, height int) *image.RGBA {
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	xDraw.CatmullRom.Scale(resized, resized.Bounds(), source, source.Bounds(), xDraw.Src, nil)
	return resized
}
