package gofpdf

import (
	//"fmt"
	"regexp"
	"strings"
)

type CssElemet map[string]string

type CssDef map[string]CssElemet

func parseSvgStyles(cssArr []string) (styles CssDef) {
	var (
		elemDef CssElemet
	)
	styles = make(CssDef)
	re := regexp.MustCompile(`([\*|\w|\-|\.]+)\s*\{(.*)\}`)
	for _, css := range cssArr {
		for _, s := range re.FindAllStringSubmatch(css, -1) {
			class := strings.TrimSpace(s[1])
			//class = strings.Replace(class, "*.", "", -1)
			cssContent := s[2]
			if elemDef = styles[class]; elemDef == nil {
				elemDef = make(CssElemet)
			}
			for _, cssElems := range strings.Split(cssContent, ";") {
				if len(cssElems) == 0 {
					continue
				}
				cssElem := strings.Split(cssElems, ":")
				elemDef[strings.TrimSpace(cssElem[0])] = strings.TrimSpace(cssElem[1])
			}
			styles[class] = elemDef
		}
	}
	return
}

func (s SVGBasicType) getStyle(className string) (elem CssElemet) {

	if elem = s.Styles[className]; elem != nil {
		return
	}
	return s.Styles["*."+className]
}
