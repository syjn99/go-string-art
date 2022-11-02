package main

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/fogleman/gg"
)

type Coordinate struct {
	X, Y float64
}

const (
	IMG_SIZE        = 32
	PIN_NUMBER      = 8
	MAX_THREAD_NUM  = 100
	MIN_DISTANCE    = 1
	BLACKNESS_DELTA = 10
	IMG_PATH        = "./sam.png"
	WHITEST         = 256
)

// int must be replaced by other custome types
var (
	pinSequence []int
	pixels      [][]int
	blackness   [][]int
	PinCoord    []Coordinate
)

func main() {
	// 1. get image size IMG_SIZE * IMG_SIZE

	img, _ := gg.LoadImage(IMG_PATH)

	// 2. convert it to gray scale. And save the pixel data

	for i := 0; i < IMG_SIZE; i++ {
		pixelsY := make([]int, 0)
		for j := 0; j < IMG_SIZE; j++ {
			rgbaColor, ok := color.RGBAModel.Convert(img.At(i, j)).(color.RGBA)
			if !ok {
				fmt.Println("color.color conversion went wrong")
			}
			grey := int(float64(rgbaColor.R)*0.21 + float64(rgbaColor.G)*0.72 + float64(rgbaColor.B)*0.07)
			pixelsY = append(pixelsY, grey)
		}
		pixels = append(pixels, pixelsY)
	}

	// 3. calculate pin coordinates via given PIN_NUMBER
	// 		use cos, sin func in math lib
	//		store pin's index with it.
	PinCoord = make([]Coordinate, PIN_NUMBER)
	calculatePins()

	// 4. calculate each line's 'blackness(if the number is lower, more black)' by averaging the pixel's Y value on the line.
	//		store the blackness info in the 2D slice [i][j]int: i for startpoint, j for endpoint. if i == j, then 256.
	// 4a. use goroutines, max numgoroutine = 10 -> goroutine is slow
	// 4b. use bresenham algorithm for getting the blackness for each line.
	blackness = make([][]int, PIN_NUMBER)

	for i := range blackness {
		blackness[i] = make([]int, PIN_NUMBER)
	}

	startTime := time.Now()
	calculateBlackness()
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("Taken, " + strconv.FormatFloat(diff.Seconds(), 'f', 6, 64))

	for i, bl := range blackness {
		fmt.Println(i, bl)
	}

	// 5. loop while step < MAX_THREAD_NUM
	// 5a. starting at pin index 0, find the 'blackest' endpoint by searching the blackness slice. the endpoint must not be the prev startpoint, and the distance must be greater than MIN_DISTANCE from startpoint and prev startpoint
	// 5b. append the endpoint index at the slice(pinSequence). Then reduce the blackness of the pixel data by BLACKNESS_DELTA. (pixel's data >= 0). Recalculate the blackness map.
	// 5c. step++.

	// 6. Given pinSequence, draw the string art.
	// 6a. using graphic lib in golang, draw line step by step. The pinSequence index => get Coordinate data => lining.
	// 6b. make the new Rect by 'result.jpg'.

	dc := gg.NewContext(1024, 1024)
	dc.SetHexColor("fff")
	dc.Clear()

	// Draw the points
	for _, coord := range PinCoord {
		dc.DrawCircle(coord.X, coord.Y, 5)
		dc.SetRGBA(0, 0, 0, 1)
		dc.Fill()
	}

	// Draw the lines
	// pinSequence = []int{0, 18, 15, 12, 6, 3}

	for _, index := range pinSequence {
		coord := PinCoord[index]
		dc.LineTo(coord.X, coord.Y)
	}
	dc.Stroke()

	dc.SavePNG("out.png")

}

func calculatePins() {
	center, radius := float64(IMG_SIZE/2), float64(IMG_SIZE/2-1)

	for i := 0; i < PIN_NUMBER; i++ {
		angle := 2 * math.Pi * float64(i) / float64(PIN_NUMBER)
		PinCoord[i] = Coordinate{
			X: math.Floor(center + radius*math.Cos(angle)),
			Y: math.Floor(center - radius*math.Sin(angle)),
		}
	}
}

func calculateBlackness() {
	for i := 0; i < PIN_NUMBER; i++ {
		for j := 0; j < PIN_NUMBER; j++ {
			if blackness[j][i] != 0 {
				blackness[i][j] = blackness[j][i]
			} else if isDistant(i, j) {
				trace := getLineTrace(i, j)
				pixelBlack := make([]int, len(trace))
				for i, coord := range trace {
					pixelBlack[i] = pixels[int(coord.X)][int(coord.Y)]
				}
				blackness[i][j] = int(average(pixelBlack))
			} else {
				blackness[i][j] = WHITEST
			}
		}
	}
}

func calculateBlacknessGoroutine() {
	var wg sync.WaitGroup

	for i := 0; i < PIN_NUMBER; i++ {
		for j := 0; j < PIN_NUMBER; j++ {
			wg.Add(1)
			go func(i, j int) {
				if blackness[j][i] != 0 {
					blackness[i][j] = blackness[j][i]
				} else if isDistant(i, j) {
					trace := getLineTrace(i, j)
					pixelBlack := make([]int, len(trace))
					for i, coord := range trace {
						pixelBlack[i] = pixels[int(coord.X)][int(coord.Y)]
					}
					blackness[i][j] = int(average(pixelBlack))
				} else {
					blackness[i][j] = WHITEST
				}
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()
}

// i: startpoint, j: endpoint
func reduceBlackness(i, j int) {
}

// return the slice of Coordinate from source point(i) to target point(j)
func getLineTrace(i, j int) []Coordinate {
	sourceCoord, targetCoord := PinCoord[i], PinCoord[j]
	dx, dy := targetCoord.X-sourceCoord.X, targetCoord.Y-sourceCoord.Y
	xsign, ysign := -1, -1
	if dx > 0 {
		xsign = 1
	}
	if dy > 0 {
		ysign = 1
	}

	dx, dy = math.Abs(dx), math.Abs(dy)

	xx, xy, yx, yy := 0, 0, 0, 0

	if dx > dy {
		xx, xy, yx, yy = xsign, 0, 0, ysign
	} else {
		xx, xy, yx, yy = 0, ysign, xsign, 0
	}

	D, y := 2*dy-dx, 0

	trace := make([]Coordinate, int(dx)+1)

	for x := range trace {
		X := sourceCoord.X + float64(x*xx+y*yx)
		Y := sourceCoord.Y + float64(x*xy+y*yy)
		trace[x] = Coordinate{X, Y}
		if D > 0 {
			y++
			D -= dx
		}
		D += dy
	}

	return trace
}

// return average value of give int slice
func average(xs []int) float64 {
	total := 0
	for _, v := range xs {
		total += v
	}
	return float64(total) / float64(len(xs))
}

// return true if the distance from i to j is enoughly distant (>= MIN_DISTANCE)
func isDistant(i, j int) bool {
	diff := int(math.Abs(float64(i - j)))
	return (diff >= MIN_DISTANCE) && (PIN_NUMBER-diff >= MIN_DISTANCE)
}
