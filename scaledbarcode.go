package barcode

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
)

type wrapFunc func(x, y int) color.Color

type scaledBarcode struct {
	wrapped     Barcode
	wrapperFunc wrapFunc
	rect        image.Rectangle
}

type intCSscaledBC struct {
	scaledBarcode
}

func (bc *scaledBarcode) Content() string {
	return bc.wrapped.Content()
}

func (bc *scaledBarcode) Metadata() Metadata {
	return bc.wrapped.Metadata()
}

func (bc *scaledBarcode) ColorModel() color.Model {
	return bc.wrapped.ColorModel()
}

func (bc *scaledBarcode) Bounds() image.Rectangle {
	return bc.rect
}

func (bc *scaledBarcode) At(x, y int) color.Color {
	return bc.wrapperFunc(x, y)
}

func (bc *intCSscaledBC) CheckSum() int {
	if cs, ok := bc.wrapped.(BarcodeIntCS); ok {
		return cs.CheckSum()
	}
	return 0
}

// Scale returns a resized barcode with the given width and height.
func Scale(bc Barcode, width, height, offset int) (Barcode, error) {
	switch bc.Metadata().Dimensions {
	case 1:
		return scale1DCode(bc, width, height, offset)
	case 2:
		return scale2DCode(bc, width, height, offset)
	}

	return nil, errors.New("unsupported barcode format")
}

func newScaledBC(wrapped Barcode, wrapperFunc wrapFunc, rect image.Rectangle) Barcode {
	result := &scaledBarcode{
		wrapped:     wrapped,
		wrapperFunc: wrapperFunc,
		rect:        rect,
	}

	if _, ok := wrapped.(BarcodeIntCS); ok {
		return &intCSscaledBC{*result}
	}
	return result
}

func scale2DCode(bc Barcode, width, height, offset int) (Barcode, error) {
	orgBounds := bc.Bounds()
	orgWidth := orgBounds.Max.X - orgBounds.Min.X
	orgHeight := orgBounds.Max.Y - orgBounds.Min.Y

	factor := int(math.Min(float64(width - 2 * offset)/float64(orgWidth), float64(height - 2 * offset)/float64(orgHeight)))
	if factor <= 0 {
		return nil, fmt.Errorf("can not scale barcode to an image smaller than %dx%d", orgWidth, orgHeight)
	}

	wrap := func(x, y int) color.Color {
		if x < offset || y < offset {
			return color.White
		}
		x = (x - offset) / factor
		y = (y - offset) / factor
		if x >= orgWidth || y >= orgHeight {
			return color.White
		}
		return bc.At(x, y)
	}

	return newScaledBC(
		bc,
		wrap,
		image.Rect(0, 0, width, height),
	), nil
}

func scale1DCode(bc Barcode, width, height, offset int) (Barcode, error) {
	orgBounds := bc.Bounds()
	orgWidth := orgBounds.Max.X - orgBounds.Min.X
	factor := float64(width - 2 * offset) / float64(orgWidth)

	if factor <= 0 {
		return nil, fmt.Errorf("can not scale barcode to an image smaller than %dx1", orgWidth)
	}

	wrap := func(x, y int) color.Color {
		if x < offset {
			return color.White
		}
		x = int(float64(x - offset) / factor)

		if x >= orgWidth {
			return color.White
		}
		return bc.At(x, 0)
	}

	return newScaledBC(
		bc,
		wrap,
		image.Rect(0, 0, width, height),
	), nil
}
