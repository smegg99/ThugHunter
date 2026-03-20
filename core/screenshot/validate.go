// core/screenshot/validate.go
package screenshot

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

// blankVarianceThreshold is the max per-channel variance for an image
// to be considered blank.
const blankVarianceThreshold = 12.0

// Validate returns true if data contains a meaningful (non-blank) image.
func Validate(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}
	return !isBlankImage(img)
}

// isBlankImageFile decodes an image file and checks for blankness.
func isBlankImageFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return false, err
	}
	return isBlankImage(img), nil
}

// isBlankImage samples up to 10 000 pixels and returns true if the
// per-channel color variance is below blankVarianceThreshold.
func isBlankImage(img image.Image) bool {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	total := w * h
	if total == 0 {
		return true
	}

	step := sampleStep(total)
	stats := samplePixels(img, step)
	return stats.avgVariance() < blankVarianceThreshold
}

// pixelStats accumulates running sums for variance calculation.
type pixelStats struct {
	sumR, sumG, sumB    float64
	sumR2, sumG2, sumB2 float64
	n                   float64
}

// avgVariance returns the mean of the per-channel variances.
func (s *pixelStats) avgVariance() float64 {
	if s.n == 0 {
		return 0
	}
	varR := (s.sumR2 / s.n) - math.Pow(s.sumR/s.n, 2)
	varG := (s.sumG2 / s.n) - math.Pow(s.sumG/s.n, 2)
	varB := (s.sumB2 / s.n) - math.Pow(s.sumB/s.n, 2)
	return (varR + varG + varB) / 3.0
}

// sampleStep returns the stride for sampling ~10 000 pixels.
func sampleStep(total int) int {
	if total > 10000 {
		return total / 10000
	}
	return 1
}

// samplePixels iterates over the image, accumulating color statistics.
func samplePixels(img image.Image, step int) pixelStats {
	bounds := img.Bounds()
	w := bounds.Dx()
	var s pixelStats

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := (y-bounds.Min.Y)*w + (x - bounds.Min.X)
			if idx%step != 0 {
				continue
			}
			r, g, b, _ := img.At(x, y).RGBA()
			rf, gf, bf := float64(r>>8), float64(g>>8), float64(b>>8)
			s.sumR += rf
			s.sumG += gf
			s.sumB += bf
			s.sumR2 += rf * rf
			s.sumG2 += gf * gf
			s.sumB2 += bf * bf
			s.n++
		}
	}
	return s
}
