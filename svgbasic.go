/*
 * Copyright (c) 2014 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package gofpdf

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var pathCmdSub *strings.Replacer

func init() {
	// Handle permitted constructions like "100L200,230"
	pathCmdSub = strings.NewReplacer(",", " ",
		"L", " L ", "l", " l ",
		"C", " C ", "c", " c ",
		"M", " M ", "m", " m ")
}

// SVGBasicSegmentType describes a single curve or position segment
type SVGBasicSegmentType struct {
	Cmd   byte // See http://www.w3.org/TR/SVG/paths.html for path command structure
	Arg   [6]float64
	Class string
	Fill  bool
}

type textType struct {
	Transform string
	Text      string
}

func (t textType) XY() (x, y float64) {
	re := regexp.MustCompile(`\d+\.\d+`)
	if sstr := re.FindAllString(t.Transform, -1); len(sstr) > 0 {
		x, _ = strconv.ParseFloat(sstr[4], 64)
		y, _ = strconv.ParseFloat(sstr[5], 64)
	}
	return

}

type srcRect struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
	W float64 `xml:"width,attr"`
	H float64 `xml:"height,attr"`
}

type pathType struct {
	D     string `xml:"d,attr"`
	Class string `xml:"class,attr"`
}
type gType struct {
	Paths []pathType `xml:"path"`
	Texts []srcText  `xml:"text"`
}
type srcText struct {
	Transform string `xml:"transform,attr"`
	Tspan     string `xml:"tspan"`
}
type srcType struct {
	ViewBox string     `xml:"viewBox,attr" `
	Rect    srcRect    `xml:"clipPath>rect"`
	Wd      float64    `xml:"width,attr"`
	Ht      float64    `xml:"height,attr"`
	Paths   []pathType `xml:"path"`
	G       []gType    `xml:"g"`
	Styles  []string   `xml:"defs>style"`
	Texts   []srcText  `xml:"text"`
}

func absolutizePath(segs []SVGBasicSegmentType) {
	var x, y float64
	var segPtr *SVGBasicSegmentType
	adjust := func(pos int, adjX, adjY float64) {
		segPtr.Arg[pos] += adjX
		segPtr.Arg[pos+1] += adjY
	}
	for j, seg := range segs {
		segPtr = &segs[j]
		if j == 0 && seg.Cmd == 'm' {
			segPtr.Cmd = 'M'
		}
		switch segPtr.Cmd {
		case 'M':
			x = seg.Arg[0]
			y = seg.Arg[1]
		case 'm':
			adjust(0, x, y)
			segPtr.Cmd = 'M'
			x = segPtr.Arg[0]
			y = segPtr.Arg[1]
		case 'L':
			x = seg.Arg[0]
			y = seg.Arg[1]
		case 'l':
			adjust(0, x, y)
			segPtr.Cmd = 'L'
			x = segPtr.Arg[0]
			y = segPtr.Arg[1]
		case 'C':
			x = seg.Arg[4]
			y = seg.Arg[5]
		case 'c':
			adjust(0, x, y)
			adjust(2, x, y)
			adjust(4, x, y)
			segPtr.Cmd = 'C'
			x = segPtr.Arg[4]
			y = segPtr.Arg[5]
		}
	}
}

func pathParse(path pathType) (segs []SVGBasicSegmentType, err error) {
	var (
		seg                             SVGBasicSegmentType
		j, argJ, argCount, prevArgCount int
		fill                            bool
	)
	pathStr := path.D
	seg.Class = path.Class
	//fmt.Printf("Path str 1 : %s\n", pathStr)
	setup := func(n int) {
		// It is not strictly necessary to clear arguments, but result may be clearer
		// to caller
		for j := 0; j < len(seg.Arg); j++ {
			seg.Arg[j] = 0.0
		}
		argJ = 0
		argCount = n
		prevArgCount = n
	}
	var str string
	var c byte
	pathStr = pathCmdSub.Replace(pathStr)
	//fmt.Printf("Path str 2 : %s\n", pathStr)
	if pLen := len(pathStr); pLen > 0 {
		fill = pathStr[pLen-1] == 'z'
	}

	strList := strings.Fields(pathStr)
	//fmt.Printf("Fields : %s\n", strList)
	count := len(strList)
	for j = 0; j < count && err == nil; j++ {
		str = strList[j]
		if argCount == 0 { // Look for path command or argument continuation
			c = str[0]
			if c == '-' || (c >= '0' && c <= '9') { // More arguments
				if j > 0 {
					setup(prevArgCount)
					// Repeat previous action
					if seg.Cmd == 'M' {
						seg.Cmd = 'L'
					} else if seg.Cmd == 'm' {
						seg.Cmd = 'l'
					}
				} else {
					err = fmt.Errorf("expecting SVG path command at first position, got %s", str)
				}
			}
		}
		if err == nil {
			if argCount == 0 {
				seg.Cmd = str[0]
				switch seg.Cmd {
				case 'M', 'm': // Absolute/relative moveto: x, y
					setup(2)
				case 'C', 'c': // Absolute/relative Bézier curve: cx0, cy0, cx1, cy1, x1, y1
					setup(6)
				case 'L', 'l': // Absolute/relative lineto: x, y
					setup(2)
				default:
					//err = fmt.Errorf("expecting SVG path command at position %d, got %s", j, str)
				}
			} else {
				seg.Arg[argJ], err = strconv.ParseFloat(str, 64)
				if err == nil {
					argJ++
					argCount--
					if argCount == 0 {
						segs = append(segs, seg)
					}
				}
			}
		}
	}
	if err == nil {
		if argCount == 0 {
			absolutizePath(segs)
		} else {
			err = fmt.Errorf("expecting additional (%d) numeric arguments", argCount)
		}
	}
	for i := range segs {
		segs[i].Fill = fill
	}
	return
}

// SVGBasicType aggregates the information needed to describe a multi-segment
// basic vector image
type SVGBasicType struct {
	Wd, Ht   float64
	X, Y     float64
	Segments [][]SVGBasicSegmentType
	Styles   CssDef
	Texts    []textType
}

// SVGBasicParse parses a simple scalable vector graphics (SVG) buffer into a
// descriptor. Only a small subset of the SVG standard, in particular the path
// information generated by jSignature, is supported. The returned path data
// includes only the commands 'M' (absolute moveto: x, y), 'L' (absolute
// lineto: x, y), and 'C' (absolute cubic Bézier curve: cx0, cy0, cx1, cy1,
// x1,y1).
func SVGBasicParse(buf []byte) (sig SVGBasicType, err error) {
	var src srcType
	err = xml.Unmarshal(buf, &src)
	rec := src.Rect
	fmt.Printf("x %f y %f w %f h %f\n", rec.X, rec.Y, rec.W, rec.H)
	if err == nil {
		//		if src.Wd == 0 && src.Ht == 0 && src.ViewBox != "" {
		//			ints := strings.Split(src.ViewBox, " ")
		//			src.Wd, _ = strconv.ParseFloat(ints[2], 64)
		//			src.Ht, _ = strconv.ParseFloat(ints[3], 64)
		//		}
		src.Wd = rec.W
		src.Ht = rec.H
		sig.Styles = parseSvgStyles(src.Styles)
		if src.Wd > 0 && src.Ht > 0 {
			paths := make([]pathType, 0, 40)
			texts := make([]srcText, 0, 40)
			paths = append(paths, src.Paths...)
			texts = append(texts, src.Texts...)
			for _, g := range src.G {
				paths = append(paths, g.Paths...)
				texts = append(texts, g.Texts...)
			}
			sig.Wd, sig.Ht = src.Wd, src.Ht
			sig.X, sig.Y = rec.X, rec.Y
			var segs []SVGBasicSegmentType
			//for _, path := range src.Paths {
			for _, path := range paths {
				if err == nil {
					segs, err = pathParse(path)
					if err != nil {
						fmt.Errorf("Parse path error %v\n", err)
					}
					if err == nil {
						sig.Segments = append(sig.Segments, segs)
					}
				}
			}
			for _, text := range texts {
				sig.Texts = append(sig.Texts, textType{text.Transform, text.Tspan})
			}
		} else {
			err = fmt.Errorf("unacceptable values for basic SVG extent: %.2f x %.2f",
				sig.Wd, sig.Ht)
		}
	}
	return
}

// SVGBasicFileParse parses a simple scalable vector graphics (SVG) file into a
// basic descriptor. See SVGBasicParse for additional comments and tutorial 20
// for an example of this function.
func SVGBasicFileParse(svgFileStr string) (sig SVGBasicType, err error) {
	var buf []byte
	buf, err = ioutil.ReadFile(svgFileStr)
	if err == nil {
		sig, err = SVGBasicParse(buf)
	}
	return
}
