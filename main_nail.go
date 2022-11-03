package main

import (
	"github.com/fogleman/gg"
	"math"
)

type Coordinate struct {
	X, Y float64
}

const (
	IMG_SIZE        = 1024
	PIN_NUMBER      = 300
	MAX_THREAD_NUM  = 4000
	MIN_DISTANCE    = 20
	BLACKNESS_DELTA = 50
	IMG_PATH        = "./iu_crop.jpeg"
	LINE_WIDTH      = 0.1
)

var (
	PinCoord    []Coordinate
	pinSequence []int
	pixels      [][]int
	usedLines   [][]bool
)

func main() {

	// 1. get image size IMG_SIZE * IMG_SIZE
	img, _ := gg.LoadImage(IMG_PATH)

	// 2. save the pixel data. (255 - grayscale, which means bigger num == blacker)
	for i := 0; i < IMG_SIZE; i++ {
		pixelsY := make([]int, 0)
		for j := 0; j < IMG_SIZE; j++ {
			// originalColor, ok := color.RGBAModel.Convert(img.At(i, j)).(color.RGBA)
			// if !ok {
			// }
			// grey := int(float64(originalColor.R)*0.21 + float64(originalColor.G)*0.72 + float64(originalColor.B)*0.07)

			r, _, _, _ := img.At(i, j).RGBA()
			grey := int(r / 257)
			pixelsY = append(pixelsY, 255-grey)
		}
		pixels = append(pixels, pixelsY)
	}

	// 3. calculate pin coordinates via given PIN_NUMBER
	//		store pin's index with it.
	PinCoord = make([]Coordinate, PIN_NUMBER)
	calculatePins()

	// 4. calculate line cache for faster algorithm. lineCache[i][j] contains the pixel coordinates from pin i to pin j. Used Bresenham algorithm

	// 5. loop while step < MAX_THREAD_NUM

	usedLines = make([][]bool, PIN_NUMBER)
	for i := range usedLines {
		usedLines[i] = make([]bool, PIN_NUMBER)
	}

	pinSequence = append(pinSequence, 0)
	curr := 0
	for i := 0; i < MAX_THREAD_NUM; i++ {
		lineBlackness, maxBlackness, blackestIndex := 0, 0, -1
		for j := 0; j < PIN_NUMBER; j++ {
			if isDistant(curr, j) {
				if usedLines[curr][j] || usedLines[j][curr] {
					continue
				}
				lineBlackness = 0
				trace := getLineTrace(curr, j)
				for _, coord := range trace {
					lineBlackness += pixels[int(coord.X)][int(coord.Y)]
				}
				if lineBlackness > maxBlackness {
					maxBlackness = lineBlackness
					blackestIndex = j
				}
			}
		}
		pinSequence = append(pinSequence, blackestIndex)
		usedLines[curr][blackestIndex], usedLines[blackestIndex][curr] = true, true
		trace := getLineTrace(curr, blackestIndex)
		for _, coord := range trace {
			temp := pixels[int(coord.X)][int(coord.Y)] - BLACKNESS_DELTA
			if temp < 0 {
				temp = 0
			}
			pixels[int(coord.X)][int(coord.Y)] = temp
		}
		curr = blackestIndex
	}

	// 6. Given pinSequence, draw the string art.
	// 6a. using graphic lib in golang, draw line step by step. The pinSequence index => get Coordinate data => lining.
	// 6b. make the new Rect by 'result_example.png'.
	dc := gg.NewContext(1024, 1024)
	dc.SetHexColor("fff")
	dc.Clear()

	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(LINE_WIDTH)
	for _, index := range pinSequence {
		coord := PinCoord[index]
		dc.LineTo(coord.X, coord.Y)
	}
	dc.Stroke()

	dc.SavePNG("result_example.png")
}

func calculatePins() {
	center, radius := float64(IMG_SIZE/2), float64(IMG_SIZE/2-1)
	for i := 0; i < PIN_NUMBER; i++ {
		angle := 2 * math.Pi * float64(i) / float64(PIN_NUMBER)
		PinCoord[i] = Coordinate{
			X: math.Floor(center + radius*math.Cos(angle)),
			Y: math.Floor(center + radius*math.Sin(angle)),
		}
	}
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
		dx, dy = dy, dx
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

// return true if the distance from i to j is enoughly distant (>= MIN_DISTANCE)
func isDistant(i, j int) bool {
	diff := int(math.Abs(float64(i - j)))
	return (diff >= MIN_DISTANCE) && (PIN_NUMBER-diff >= MIN_DISTANCE)
}

// return true if the slice contains num
func contains(arr []int, num int) bool {
	for i := range arr {
		if arr[i] == num {
			return true
		}
	}
	return false
}
