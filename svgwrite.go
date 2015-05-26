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
//	"fmt"
)

const (
	// FontSize is default font size
	FontSize = 4.0
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
	var (
		x, y, newX, newY   float64
		cx0, cy0, cx1, cy1 float64
		path               []SVGBasicSegmentType
		seg                SVGBasicSegmentType
		points             []PointType
	)
	originX, originY := f.GetXY()
	lineW := f.GetLineWidth()
	val := func(arg int) (float64, float64) {
		return originX + scale*seg.Arg[arg], originY + scale*seg.Arg[arg+1]
	}
	for j := 0; j < len(sb.Segments) && f.Ok(); j++ {
		path = sb.Segments[j]
		if len(path) > 0 && path[0].IsPolygon {
			points = make([]PointType, 0, 6)
		} else {
			points = nil
		}
		for k := 0; k < len(path) && f.Ok(); k++ {
			class := seg.Class
			style := sb.Styles.Get(class)
			f.SetStyle(style)

			seg = path[k]
			switch seg.Cmd {
			case 'M':
				x, y = val(0)
				f.SetXY(x, y)
			case 'L':
				newX, newY = val(0)
				l := NewLine(x, y, newX, newY)
				l.Draw(f, style)
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
			if seg.IsPolygon {
				points = append(points, PointType{x, y})
			}
			// reset line width
			f.SetLineWidth(lineW)
		}
		if points != nil {
			// don't drow white polygons
			c1, c2, c3 := f.GetFillColor()
			if c1 != 255 || c2 != 255 || c3 != 255 {
				f.Polygon(points, "F")
			}
		}
	}
}

//
// SVGWriteTexts writes SVG texts on pdf document
//
func (f *Fpdf) SVGWriteTexts(sb *SVGBasicType, scale float64) {
	for _, text := range sb.Texts {
		if len(text.Text) == 0 {
			continue
		}
		f.SVGWriteText(sb, text, scale)
	}
}

//
// SVGWriteText writes SVG text on pdf document
//
func (f *Fpdf) SVGWriteText(sb *SVGBasicType, text TextType, scale float64) {
	// set style for class
	style := sb.Styles["text."+text.Class]
	f.SetStyle(style)
	// set style for element
	f.SetStyle(text.Style)
	x, y := text.XY()
	// calc x and y shift
	shiftRatio := text.Style.BaseLineShift / 100.0
	fontSize := FontSize * text.FontScale()
	f.SetFontSize(fontSize)
	_, pointsFontSize := f.GetFontSize()
	yShift := 0.0
	for _, str := range text.Text {
		f.TransformBegin()
		tx, ty := (x * scale), ((y * scale) + yShift)
		if shiftRatio != 0 {
			textSize := f.GetStringWidth(str)
			xShift := float64(textSize * shiftRatio)
			tx += xShift
		}
		f.TransformTranslate(tx, ty)
		if text.rotation != 0 {
			f.TransformRotate(text.rotation, 0, 0)
		}
		f.Text(0, 0, str)
		f.TransformEnd()
		// if text is multiline
		yShift += pointsFontSize
	}
}

//
// SetStyle sets style color, line width etc. for current style def
//
func (f *Fpdf) SetStyle(style *StyleDef) {
	if style == nil {
		return
	}
	if style.IsStroke {
		f.SetDrawColor(style.Stroke[0], style.Stroke[1], style.Stroke[2])
	}
	if style.IsFill {
		f.SetFillColor(style.Fill[0], style.Fill[1], style.Fill[2])
		f.SetTextColor(style.Fill[0], style.Fill[1], style.Fill[2])
	}
	lineW := f.GetLineWidth()
	f.SetLineWidth(lineW * style.StrokeWidth)
}
