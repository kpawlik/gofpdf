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
	"encoding/hex"
	//	"fmt"
	"strconv"
	"strings"
)

const (
	LINE_WIDTH = 0.1
	FONT_SIZE  = 4.0
)

// SVGBasicWrite renders the paths encoded in the basic SVG image specified by
// sb. The scale value is used to convert the coordinates in the path to the
// unit of measure specified in New(). The current position (as set with a call
// to SetXY()) is used as the origin of the image. The current line cap style
// (as set with SetLineCapStyle()), line width (as set with SetLineWidth()),
// and draw color (as set with SetDrawColor()) are used in drawing the image
// paths.
//
// See example 20 for a demonstration of this function.
func (f *Fpdf) SVGBasicWrite(sb *SVGBasicType, scale float64) {
	originX, originY := f.GetXY()
	var (
		x, y, newX, newY   float64
		cx0, cy0, cx1, cy1 float64
		path               []SVGBasicSegmentType
		seg                SVGBasicSegmentType
		points             []PointType
		prevVals           map[string]string
	)
	prevVals = make(map[string]string)
	val := func(arg int) (float64, float64) {
		return originX + scale*seg.Arg[arg], originY + scale*seg.Arg[arg+1]
	}
	for j := 0; j < len(sb.Segments) && f.Ok(); j++ {
		path = sb.Segments[j]
		if len(path) > 0 && path[0].Fill {
			points = make([]PointType, 0, 6)
		} else {
			points = nil
		}
		for k := 0; k < len(path) && f.Ok(); k++ {
			f.SetLineWidth(LINE_WIDTH)
			class := seg.Class
			style := sb.getStyle(class)
			f.SetStyle(style, prevVals)
			//			if stroke := style["stroke"]; stroke != "" && stroke != prevStroke && stroke != "none" {
			//				color, _ := hex.DecodeString(strings.Replace(stroke, "#", "", -1))
			//				f.SetDrawColor(int(color[0]), int(color[1]), int(color[2]))
			//				prevStroke = stroke
			//			}
			//			if strokeWidth := style["stroke-width"]; strokeWidth != prevStrokeWidth {
			//				w, _ := strconv.ParseFloat(strings.Replace(strokeWidth, "px", "", -1), 32)
			//				f.SetLineWidth(LINE_WIDTH * w)
			//				prevStrokeWidth = strokeWidth
			//			}
			//			if fill := style["fill"]; fill != "" && fill != prevFill && fill != "none" {
			//				color, _ := hex.DecodeString(strings.Replace(fill, "#", "", -1))
			//				f.SetFillColor(int(color[0]), int(color[1]), int(color[2]))
			//				prevFill = fill
			//			}

			seg = path[k]
			switch seg.Cmd {
			case 'M':
				x, y = val(0)
				f.SetXY(x, y)
			case 'L':
				newX, newY = val(0)
				f.Line(x, y, newX, newY)
				x, y = newX, newY
			case 'C':
				cx0, cy0 = val(0)
				cx1, cy1 = val(2)
				newX, newY = val(4)
				f.CurveCubic(x, y, cx0, cy0, newX, newY, cx1, cy1, "D")
				x, y = newX, newY
			default:
				f.SetErrorf("Unexpected path command '%c'", seg.Cmd)
			}
			if seg.Fill {
				points = append(points, PointType{x, y})
			}
		}
		if points != nil {
			c1, c2, c3 := f.GetFillColor()
			if c1 != 255 || c2 != 255 || c3 != 255 {
				f.Polygon(points, "F")
			}
		}
	}
}

func (p *Fpdf) SVGTextWrite(sb *SVGBasicType, scale float64) {
	prevVals := make(map[string]string)
	for _, text := range sb.Texts {
		if len(text.Text) == 0 {
			continue
		}
		//fontSize := FONT_SIZE * text.FontScale()

		p.SetStyle(text.Style, prevVals)
		//fmt.Println(text.Class)
		style := sb.Styles["text."+text.Class]
		str := text.Text
		//fmt.Printf("%s\n  %v\n  %v\n", str, text.Style, style)
		p.SetStyle(style, prevVals)
		x, y := text.XY()

		p.SetFontSize(FONT_SIZE * text.FontScale())
		tx, ty := x*scale, y*scale
		if text.Style["text-anchor"] == "middle" {
			textSize := p.GetStringWidth(str)
			tx -= textSize / 2
		}
		p.TransformBegin()
		p.TransformTranslate(tx, ty)
		if text.rotation != 0 {
			p.TransformRotate(text.rotation, 0, 0)
		}
		p.Text(0, 0, str)
		p.TransformEnd()
	}
}

func (f *Fpdf) SetStyle(style CssElemet, prevVals map[string]string) {
	var (
		key string
	)
	//style := sb.Styles[class]
	key = "stroke"
	if stroke, found := style[key]; found && stroke != "none" {
		if prevStroke := prevVals[key]; prevStroke != stroke {
			color, _ := hex.DecodeString(strings.Replace(stroke, "#", "", -1))
			f.SetDrawColor(int(color[0]), int(color[1]), int(color[2]))
			prevVals[key] = stroke
		}
	}
	key = "stroke-width"
	if strokeWidth, found := style[key]; found {
		if prevStrokeWidth := prevVals[key]; strokeWidth != prevStrokeWidth {
			w, _ := strconv.ParseFloat(strings.Replace(strokeWidth, "px", "", -1), 32)
			f.SetLineWidth(LINE_WIDTH * w)
			prevVals[key] = prevStrokeWidth
		}
	}
	key = "fill"
	if fill, found := style[key]; found && fill != "none" {
		if prevFill := prevVals[key]; prevFill != fill {
			color, _ := hex.DecodeString(strings.Replace(fill, "#", "", -1))
			f.SetFillColor(int(color[0]), int(color[1]), int(color[2]))
			f.SetTextColor(int(color[0]), int(color[1]), int(color[2]))
			prevVals[key] = prevFill
		}
	}
}
