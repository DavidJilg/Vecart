package general

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
)

type RGB struct {
	red, green, blue uint8
}

func (rgb *RGB) toString() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.red, rgb.green, rgb.blue)
}

func (rgb *RGB) fromString(hex string){
    red, err := strconv.ParseUint(hex[1:3], 16, 64)
    if err != nil{
        fmt.Printf("\nCould not convert '%s' to RGB | %e", hex, err)
        return
    }
    rgb.red = uint8(red)

    green, err := strconv.ParseUint(hex[3:5], 16, 64)
    if err != nil{
        fmt.Printf("\nCould not convert '%s' to RGB | %e", hex, err)
        return
    }
    rgb.green = uint8(green)

    blue, err := strconv.ParseUint(hex[5:7], 16, 64)
    if err != nil{
        fmt.Printf("\nCould not convert '%s' to RGB | %e", hex, err)
        return
    }
    rgb.blue = uint8(blue)
}

func (rgb *RGB) fromRGBA(color color.NRGBA){
    rgb.red = ((1 - color.A) * 255) + (color.A * color.R)
    rgb.green = ((1 - color.A) * 255) + (color.A * color.G)
    rgb.blue = ((1 - color.A) * 255) + (color.A * color.B)
}

func NewRBGFromNRGBA(color color.NRGBA) RGB{
    rgb := RGB{0,0,0}
    rgb.fromRGBA(color)
    return rgb
}

func (rgb *RGB) distance(otherRgb *RGB) float64 {
    redMean := ( int(rgb.red) + int(otherRgb.red )) / 2
    redDelta := int(rgb.red) - int(otherRgb.red)
    greenDelta := int(rgb.green) - int(otherRgb.green)
    blueDelta := int(rgb.blue) - int(otherRgb.blue)

    t1 := (((512+redMean)*redDelta*redDelta)>>8) + (4*greenDelta*greenDelta) + (((767-redMean)*blueDelta*blueDelta)>>8)

    return math.Sqrt(float64(t1))
}