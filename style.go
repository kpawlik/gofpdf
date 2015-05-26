package gofpdf

import (
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var (
	nbrRe *regexp.Regexp = regexp.MustCompile(`\-{0,1}\d+`)
)

type StyleDef struct {
	StringMap     map[string]string
	DashArray     []float64
	Stroke        [3]int
	Fill          [3]int
	IsFill        bool
	IsStroke      bool
	StrokeWidth   float64
	BaseLineShift float64
}

//
// Creates new empty StyleDef record
//
func NewEmptyStyleDef() *StyleDef {
	sm := make(map[string]string)
	return &StyleDef{StringMap: sm, StrokeWidth: 1.0}
}

//
// Returns new instance of StyleDef from string
//
func NewStyleDef(str string) *StyleDef {
	ce := NewEmptyStyleDef()
	ce.Append(str)
	return ce
}

//
// Adds attribute to style def
//
func (ce *StyleDef) Append(str string) {
	for _, cssElems := range strings.Split(str, ";") {
		if len(cssElems) == 0 {
			continue
		}
		cssElem := strings.Split(cssElems, ":")
		ce.Set(strings.TrimSpace(cssElem[0]), strings.TrimSpace(cssElem[1]))
	}
}

//
// Getter for style def
//
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
	}

}

//
// Returns true if Style definition contains attribute equal to value
//
func (ce StyleDef) Check(key, value string) bool {
	if val, ok := ce.StringMap[key]; !ok {
		return false
	} else {
		return val == value
	}

}

//*************************
//
// Stores styles for classes
//
type StylesDef map[string]*StyleDef

func NewEmptyStylesDef() StylesDef {
	return make(StylesDef)
}

func NewStylesDef(styles []string) StylesDef {
	stylesDef := NewEmptyStylesDef()
	re := regexp.MustCompile(`([\*|\w|\-|\.]+)\s*\{(.*)\}`)
	for _, style := range styles {
		for _, s := range re.FindAllStringSubmatch(style, -1) {
			class := strings.TrimSpace(s[1])
			styleContent := s[2]
			elemDef := stylesDef.Get(class)
			elemDef.Append(styleContent)
		}
	}
	return stylesDef
}

//
// Returns existing StyleDef from key. If does not exists then
// creates new one add with key and returns.
// Search order:
//  1. search for class
//  2. search for "*."+class
//  3. return empty style def
//
func (cd StylesDef) Get(class string) *StyleDef {
	if style, ok := cd[class]; ok {
		return style
	}
	if style, ok := cd["*."+class]; ok {
		return style
	}
	style := NewEmptyStyleDef()
	cd[class] = style
	return style
}
