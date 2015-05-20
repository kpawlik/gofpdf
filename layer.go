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

// Routines in this file are translated from
// http://www.fpdf.org/en/script/script97.php

import (
//"fmt"
)

type layerType struct {
	name     string
	visible  bool
	objNum   int // object number
	children []int
}

type layerRecType struct {
	list          []layerType
	currentLayer  int
	openLayerPane bool
}

func (f *Fpdf) layerInit() {
	f.layer.list = make([]layerType, 0)
	f.layer.currentLayer = -1
	f.layer.openLayerPane = false
}

// AddLayer defines a layer that can be shown or hidden when the document is
// displayed. name specifies the layer name that the document reader will
// display in the layer list. visible specifies whether the layer will be
// initially visible. The return value is an integer ID that is used in a call
// to BeginLayer().
//
// Layers are demonstrated in tutorial 26.
func (f *Fpdf) AddLayer(name string, visible bool) (layerID int) {
	layerID = len(f.layer.list)
	f.layer.list = append(f.layer.list, layerType{name: name, visible: visible, children: make([]int, 0)})
	return
}

// AddChild defines a layer to other later as child layer. Parent layer will become
// folder of layers.
//
// Layers are demonstrated in tutorial 26.
func (f *Fpdf) AddChild(parentId int, childId int) {
	l := f.layer.list[parentId]
	l.children = append(l.children, childId)
	f.layer.list[parentId] = l

}

// BeginLayer is called to begin adding content to the specified layer. All
// content added to the page between a call to BeginLayer and a call to
// EndLayer is added to the layer specified by id. See AddLayer for more
// details.
func (f *Fpdf) BeginLayer(id int) {
	f.EndLayer()
	if id >= 0 && id < len(f.layer.list) {
		f.outf("/OC /OC%d BDC", id)
		f.layer.currentLayer = id
	}
}

// EndLayer is called to stop adding content to the currently active layer. See
// BeginLayer for more details.
func (f *Fpdf) EndLayer() {
	if f.layer.currentLayer >= 0 {
		f.out("EMC")
		f.layer.currentLayer = -1
	}
}

// OpenLayerPane advises the document reader to open the layer pane when the
// document is initially displayed.
func (f *Fpdf) OpenLayerPane() {
	f.layer.openLayerPane = true
}

func (f *Fpdf) layerEndDoc() {
	if len(f.layer.list) > 0 {
		if f.pdfVersion < "1.5" {
			f.pdfVersion = "1.5"
		}
	}
}

func (f *Fpdf) layerPutLayers() {
	for j, l := range f.layer.list {
		f.newobj()
		f.layer.list[j].objNum = f.n
		f.outf("<</Type /OCG /Name %s>>", f.textstring(utf8toutf16(l.name)))
		f.out("endobj")
	}
}

func (f *Fpdf) layerPutResourceDict() {
	if len(f.layer.list) > 0 {
		f.out("/Properties <<")
		for j, layer := range f.layer.list {
			f.outf("/OC%d %d 0 R", j, layer.objNum)
		}
		f.out(">>")
	}

}

func getChildrenList(f *Fpdf) (children map[int]bool) {
	children = make(map[int]bool)
	for _, layer := range f.layer.list {
		for _, childId := range layer.children {
			child := f.layer.list[childId]
			children[child.objNum] = true
		}
	}
	return
}

func (f *Fpdf) layerPutCatalog() {
	var (
		onStr, ordStr, offStr string
	)
	if len(f.layer.list) > 0 {
		children := getChildrenList(f)
		for _, layer := range f.layer.list {
			//fmt.Println(layer.name, "  ", layer.children)
			if !children[layer.objNum] {
				ordStr += sprintf("%d 0 R ", layer.objNum)
			}
			onStr += sprintf("%d 0 R ", layer.objNum)
			if len(layer.children) > 0 {
				ordStr += " ["
				for _, childId := range layer.children {
					child := f.layer.list[childId]
					ordStr += sprintf("%d 0 R ", child.objNum)
				}
				ordStr += "] "
			}
			if !layer.visible {
				offStr += sprintf("%d 0 R ", layer.objNum)
			}
		}
		f.outf("/OCProperties <</OCGs [%s] /D << /OFF [%s] /Order [%s]>>>>", onStr, offStr, ordStr)
		if f.layer.openLayerPane {
			f.out("/PageMode /UseOC")
		}
	}
}
