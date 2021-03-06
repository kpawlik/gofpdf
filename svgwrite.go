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
		polygon            []PointType
		polygonStyle       *StyleDef
	)
	originX, originY := f.GetXY()
	lineW := f.GetLineWidth()
	val := func(arg int) (float64, float64) {
		return originX + scale*seg.Arg[arg], originY + scale*seg.Arg[arg+1]
	}
	for j := 0; j < len(sb.Segments) && f.Ok(); j++ {
		path = sb.Segments[j]
		if len(path) > 0 && path[0].IsPolygon {
			polygon = make([]PointType, 0, 6)
		} else {
			polygon = nil
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
				polygon = append(polygon, PointType{x, y})
				// cache style for polygon
				polygonStyle = style
			}
			// reset line width
			f.SetLineWidth(lineW)
		}
		if polygon != nil {
			// don't drow white polygons
			c1, c2, c3 := f.GetFillColor()
			if c1 != 255 || c2 != 255 || c3 != 255 {
				f.SetStyle(polygonStyle)
				f.Polygon(polygon, "F")
			}
		}
	}
}

// SVGWriteTexts writes SVG texts on PDF document
func (f *Fpdf) SVGWriteTexts(sb *SVGBasicType, scale float64) {
	for _, text := range sb.Texts {
		if len(text.Text) != 0 {
			f.SVGWriteText(sb, text, scale)
		}
	}
}

// SVGWriteText writes, transform and rotate SVG text on PDF document
func (f *Fpdf) SVGWriteText(sb *SVGBasicType, text TextType, scale float64) {
	// merge elemet style with class style
	style := text.Style
	style.Extend(sb.Styles.Get(text.Class))
	// set style for element
	f.SetStyle(style)
	x, y := text.XY()
	// calc x and y shift
	shiftRatio := text.Style.BaseLineShift / 100.0

	fontSize := style.FontSize * text.FontScale()
	f.SetFontSize(fontSize)
	_, pointsFontSize := f.GetFontSize()
	yShift := 0.0
	for _, str := range text.Text {
		xShift := 0.0
		if shiftRatio != 0 {
			textSize := f.GetStringWidth(str)
			xShift = textSize * shiftRatio
		}
		f.TransformBegin()
		tx, ty := ((x * scale) + xShift), ((y * scale) + yShift)
		f.TransformTranslate(tx, ty)
		if text.rotation != 0 {
			f.TransformRotate(text.rotation, 0-xShift, 0-yShift)
		}
		f.Text(0, 0, str)
		f.TransformEnd()
		// if text is multiline
		yShift += pointsFontSize
	}
}

// SetStyle sets style (color, line width etc.) for current PDF file
// from StyleDef
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
	if style.IsBold {
		f.SetFont("", "B", 0)
	} else {
		f.SetFont("", "", 0)
	}
	// Set opacity
	f.SetAlpha(style.Opacity, "Normal")
	// set line width
	lineW := f.GetLineWidth()
	f.SetLineWidth(lineW * style.StrokeWidth)
}
