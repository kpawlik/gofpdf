package gofpdf

import (
	"encoding/hex"
	//	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	nbrRe *regexp.Regexp
)

func init() {
	nbrRe = regexp.MustCompile(`\-{0,1}\d+`)
}

//StyleDef aggregates data about element style
type StyleDef struct {
	StringMap     map[string]string
	DashArray     []float64
	Stroke        [3]int
	Fill          [3]int
	IsFill        bool
	IsStroke      bool
	StrokeWidth   float64
	BaseLineShift float64
	IsBold        bool
}

//
// NewEmptyStyleDef creates new empty StyleDef record
func NewEmptyStyleDef() *StyleDef {
	sm := make(map[string]string)
	return &StyleDef{StringMap: sm, StrokeWidth: 1.0}
}

//
// NewStyleDef returns new instance of StyleDef from string
func NewStyleDef(str string) *StyleDef {
	ce := NewEmptyStyleDef()
	ce.Append(str)
	return ce
}

//
// Append adds attribute to style def
func (ce *StyleDef) Append(str string) {
	for _, cssElems := range strings.Split(str, ";") {
		if len(cssElems) == 0 {
			continue
		}
		cssElem := strings.Split(cssElems, ":")
		ce.Set(strings.TrimSpace(cssElem[0]), strings.TrimSpace(cssElem[1]))
	}
}

// Extend cureent styledef about data from another styledef
func (ce *StyleDef) Extend(style *StyleDef) {
	if ce == nil {
		return
	}
	for k, v := range style.StringMap {
		ce.Set(k, v)
	}

}

//
// Get is a getter for style def
func (ce *StyleDef) Get(key string) (string, bool) {
	val, ok := ce.StringMap[key]
	return val, ok
}

//
// Set and parse style attribute
//
func (ce *StyleDef) Set(key, value string) {
	ce.StringMap[key] = value
	switch key {
	case "stroke":
		color, _ := hex.DecodeString(strings.Replace(value, "#", "", -1))
		ce.Stroke = [...]int{int(color[0]), int(color[1]), int(color[2])}
		ce.IsStroke = true
	case "stroke-width":
		w, _ := strconv.ParseFloat(strings.Replace(value, "px", "", -1), 32)
		ce.StrokeWidth = w
	case "fill":
		if value != "none" {
			color, _ := hex.DecodeString(strings.Replace(value, "#", "", -1))
			ce.Fill = [...]int{int(color[0]), int(color[1]), int(color[2])}
			ce.IsFill = true
		}
	case "stroke-dasharray":
		for _, str := range nbrRe.FindAllString(value, -1) {
			f, _ := strconv.ParseFloat(str, 64)
			ce.DashArray = append(ce.DashArray, f)
		}
	case "baseline-shift":
		if str := nbrRe.FindString(value); len(str) > 0 {
			shift, _ := strconv.ParseFloat(str, 64)
			ce.BaseLineShift = shift
		}
	case "font-weight":
		ce.IsBold = value == "bold"
	}

}

// Check returns true if Style definition contains attribute equal to value
func (ce StyleDef) Check(key, value string) bool {
	var (
		val string
		ok  bool
	)
	if val, ok = ce.StringMap[key]; !ok {
		return false
	}
	return val == value
}

//
// StylesDef stores styles for classes
type StylesDef map[string]*StyleDef

// NewEmptyStylesDef create new empty styles def
func NewEmptyStylesDef() StylesDef {
	return make(StylesDef)
}

// NewStylesDef create new empty styles def from strings
func NewStylesDef(styles []string) StylesDef {

	stylesDef := NewEmptyStylesDef()
	re := regexp.MustCompile(`([\*|\w|\-|\.]+)\s*\{(.*)\}`)
	for _, style := range styles {
		for _, s := range re.FindAllStringSubmatch(style, -1) {
			shortClass := ""
			class := strings.TrimSpace(s[1])
			if parts := strings.Split(class, "."); len(parts) == 2 {
				shortClass = parts[1]
			}
			styleContent := s[2]
			elemDef := stylesDef.Get(class)
			elemDef.Append(styleContent)
			if len(shortClass) > 0 {
				stylesDef[shortClass] = elemDef
			}
		}
	}
	return stylesDef
}

//
// Get returns existing StyleDef from key. If does not exists then
// creates new one add with key and returns.
// Search order:
//  1. search for class
//  2. search for "*."+class
//  3. search for "text."+class
//  3. return empty style def
//
func (cd StylesDef) Get(class string) *StyleDef {
	if len(class) == 0 {
		return nil
	}

	if style, ok := cd[class]; ok {
		return style
	}
	if style, ok := cd["*."+class]; ok {
		return style
	}
	if style, ok := cd["text."+class]; ok {
		return style
	}
	style := NewEmptyStyleDef()
	cd[class] = style
	return style
}
