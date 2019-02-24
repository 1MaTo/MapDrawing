package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"

	"os"

	"github.com/davvo/mercator"
	"github.com/fogleman/gg"
	"github.com/paulmach/go.geojson"
	yaml "gopkg.in/yaml.v2"
	//"github.com/im7mortal/UTM"
)

const (
	correct        = -1000.0
	pointR         = 1
	picWidth       = 1366
	picHeight      = 1024
	scale          = 1.0
	stylesName     = "style.yml"
	defFillColor   = "#757575AA"
	defBorderColor = "#0f0f0fFF"
	defLineWidth   = 0.5
)

var fileName *string

type style struct {
	AdLevel []adminLevel `yaml:"admin_level"`
	Lines   lines        `yaml:"lines"`
}

type adminLevel struct {
	Rank        int     `yaml:"rank"`
	BorderColor string  `yaml:"borderColor"`
	FillColor   string  `yaml:"fillColor"`
	LineWidth   float64 `yaml:"lineWidth"`
}

type lines struct {
	Road   string  `yaml:"road"`
	RWidth float64 `yaml:"roadWidth"`
}

func main() {
	fileName = flag.String("path", "map.geojson", "Choose GeoJSON file")
	flag.Parse()

	fmt.Printf("\nПроверка, имя указанного файла: %s\n", *fileName)

	file, _ := os.Open(*fileName)
	defer file.Close()
	stat, _ := file.Stat()
	buf := make([]byte, stat.Size())
	file.Read(buf)

	fc, _ := geojson.UnmarshalFeatureCollection(buf)

	styles := readStyles(stylesName)

	dc := gg.NewContext(picWidth, picHeight)
	dc.SetFillRule(gg.FillRuleEvenOdd)
	dc.InvertY()

	for _, feature := range fc.Features {

		fmt.Printf("\nType: %s\n", feature.Geometry.Type)

		switch feature.Geometry.Type {
		case "Polygon":
			for _, geom := range feature.Geometry.Polygon {
				for _, point := range geom {
					x := 0.0
					if point[0] < -167.0 {
						point[0] = point[0] + 360
					}
					x, y := mercator.LatLonToPixels(point[1], point[0], 1)
					x, y = scalePic(x, y)
					dc.LineTo(x, y)
				}
				dc.ClosePath()
				setStyle(dc, styles, feature.Properties)
			}
		case "LineString":
			for _, geom := range feature.Geometry.LineString {
				x, y := mercator.LatLonToPixels(geom[1], geom[0], 1)
				x, y = scalePic(x, y)
				dc.LineTo(x, y)
			}
			setLineStyle(dc, styles, feature.Properties)
		case "Point":
			x, y := mercator.LatLonToPixels(feature.Geometry.Point[1], feature.Geometry.Point[0], 1)
			x, y = scalePic(x, y)
			dc.DrawPoint(x, y, pointR)
			setStyle(dc, styles, feature.Properties)
		case "MultiPolygon":
			for _, geom := range feature.Geometry.MultiPolygon {
				for _, poly := range geom {
					newSubPoly := true
					for _, point := range poly {
						if point[0] < -167.0 {
							point[0] = point[0] + 360
						}
						x, y := mercator.LatLonToPixels(point[1], point[0], 3)
						x, y = scalePic(x, y)
						if newSubPoly {
							dc.MoveTo(x, y)
							newSubPoly = false
						} else {
							dc.LineTo(x, y)
						}
					}
				}
				dc.ClosePath()
				setStyle(dc, styles, feature.Properties)
			}
		}
	}
	drawLegend(dc, styles)
	dc.SavePNG("out.png")
}

func drawLegend(dc *gg.Context, st style) {
	textX := 30.0
	textY := picHeight - 150.0
	textSize := 24.0
	shiftY := 50.0
	drawLegendBack(dc)
	drawLegendText(dc, textX, textY, textSize, "Administrative level 2   - ")
	drawInvRec(dc, textX+215, textY-20)
	drawStyle(dc, st, 2)
	drawLegendText(dc, textX, textY+shiftY, textSize, "Administrative level 3   - ")
	drawInvRec(dc, textX+215, textY-20+shiftY)
	drawStyle(dc, st, 3)
}

func drawInvRec(dc *gg.Context, x, y float64) {
	dc.InvertY()
	dc.DrawRectangle(x, y, 25.0, 25.0)
	dc.InvertY()
}

func drawLegendBack(dc *gg.Context) {
	dc.DrawRectangle(0.0, 0.0, 300.0, 200.0)
	dc.SetRGB255(255, 255, 255)
	dc.FillPreserve()
	dc.SetRGBA255(249, 188, 189, 255)
	dc.SetLineWidth(1.5)
	dc.Stroke()
}

func drawLegendText(dc *gg.Context, x, y, size float64, text string) {
	dc.LoadFontFace("./fonts/FRIZQT__.ttf", size)
	dc.InvertY()
	dc.SetRGB255(0, 0, 0)
	dc.DrawString(text, x, y)
	dc.InvertY()
}

func drawStyle(dc *gg.Context, st style, i int) {
	dc.SetHexColor(st.AdLevel[i].FillColor)
	dc.FillPreserve()
	dc.SetHexColor(st.AdLevel[i].BorderColor)
	dc.SetLineWidth(st.AdLevel[i].LineWidth)
	dc.Stroke()
}

func drawPoly(dc *gg.Context, fillColor string, borderColor string, lineWidth float64) {
	dc.SetHexColor(fillColor)
	dc.FillPreserve()
	dc.SetHexColor(borderColor)
	dc.SetLineWidth(lineWidth)
	dc.Stroke()
}

func drawLine(dc *gg.Context) {
	dc.SetRGBA(1, 0, 0, 0.5)
	dc.SetLineWidth(0.5)
	dc.Stroke()
}

func scalePic(x, y float64) (float64, float64) {
	x = x*scale + correct
	y = y*scale + correct
	return x, y
}

func drawP(dc *gg.Context) {
	dc.SetRGBA(0.35, 0.35, 0.35, 0.5)
	dc.Fill()
}

func readStyles(styleFile string) style {
	st := style{}
	file, _ := ioutil.ReadFile(styleFile)
	_ = yaml.Unmarshal(file, &st)
	fmt.Printf("--- t:\n%v\n\n", st)
	return st
}

func setStyle(dc *gg.Context, st style, prop map[string]interface{}) {
	err := prop["admin_level"]
	if err != nil {
		level, _ := strconv.Atoi(prop["admin_level"].(string))
		fColor := st.AdLevel[level].FillColor
		bColor := st.AdLevel[level].BorderColor
		lWidth := st.AdLevel[level].LineWidth
		if fColor != "" {
			dc.SetHexColor(fColor)
		} else {
			dc.SetHexColor(defFillColor)
		}
		dc.FillPreserve()
		if bColor != "" {
			dc.SetHexColor(bColor)
		} else {
			dc.SetHexColor(defBorderColor)
		}
		if lWidth != 0.0 {
			dc.SetLineWidth(lWidth)
		} else {
			dc.SetLineWidth(defLineWidth)
		}
	} else {
		dc.SetHexColor(defFillColor)
		dc.FillPreserve()
		dc.SetHexColor(defBorderColor)
		dc.SetLineWidth(defLineWidth)
	}
	dc.Stroke()
}

func setPointStyle(dc *gg.Context, st style, prop map[string]interface{}) {}

func setLineStyle(dc *gg.Context, st style, prop map[string]interface{}) {
	err := prop["road"]
	if err != nil {
		roadValue, _ := strconv.ParseBool(prop["road"].(string))
		if roadValue {
			fmt.Println("RenderRoad")
			dc.SetHexColor(st.Lines.Road)
			dc.SetLineWidth(st.Lines.RWidth)
		} else {
			dc.SetHexColor(defBorderColor)
			dc.SetLineWidth(defLineWidth)
		}
		dc.Stroke()
	}
}
