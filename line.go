package gofpdf

import (
	"errors"
	//	"fmt"
	"math"
	"regexp"
	//"strconv"
)

var (
	re *regexp.Regexp = regexp.MustCompile(`\d+`)
)

type Line struct {
	x1, y1, x2, y2 float64
}

func NewLine(x1, y1, x2, y2 float64) *Line {
	return &Line{x1, y1, x2, y2}
}

func (l Line) Length() float64 {
	return math.Sqrt(math.Pow((l.x2-l.x1), 2) + math.Pow((l.y2-l.y1), 2))
}

func (l Line) Draw(f *Fpdf, style *StyleDef) {
	if dash := style.DashArray; len(dash) == 0 {
		f.Line(l.x1, l.y1, l.x2, l.y2)
	} else {
		l.DrawDashed(f, dash)
	}
}

func (l Line) DrawDashed(f *Fpdf, dashArray []float64) {
	//
	// algorithm source: https://deepanjandas.wordpress.com/2010/06/17/draw-dashed-line-in-flash/
	//
	if len(dashArray) != 2 {
		return
	}
	stroke, gap := dashArray[0]/6, dashArray[1]/6
	segmentLength := stroke + gap
	deltaX := l.x2 - l.x1
	deltaY := l.y2 - l.y1
	delta := math.Sqrt((deltaX * deltaX) + (deltaY * deltaY))
	segmentCount := int(math.Floor(math.Abs(delta / segmentLength)))
	radians := math.Atan2(deltaY, deltaX)
	ax, ay := l.x1, l.y1
	deltaX = math.Cos(radians) * segmentLength
	deltaY = math.Sin(radians) * segmentLength
	for i := 0; i < segmentCount; i++ {
		f.Line(ax, ay, ax+math.Cos(radians)*stroke, ay+math.Sin(radians)*stroke)
		ax += deltaX
		ay += deltaY
	}
	// handle last section
	delta = math.Sqrt((l.x2-ax)*(l.x2-ax) + (l.y2-ay)*(l.y2-ay))
	if delta > segmentLength {
		f.Line(ax, ay, ax+math.Cos(radians)*stroke, ay+math.Sin(radians)*stroke)
	} else if delta > 0 {
		f.Line(ax, ay, ax+math.Cos(radians)*delta, ay+math.Sin(radians)*delta)
	}
}

func parseDash(dash string) (dashArray [2]float64, err error) {
	strArr := re.FindAllString(dash, -1)
	if len(strArr) != 2 {
		err = errors.New("Canot parse dash array for line")
		return
	}
	//dashArray[0] = strconv
	return

}
