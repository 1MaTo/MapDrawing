package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"

	"os"

	"github.com/fogleman/gg"
	"github.com/paulmach/go.geojson"
	yaml "gopkg.in/yaml.v2"
	//"github.com/im7mortal/UTM"
)

const (
	correct        = 0
	pointR         = 1
	picWidth       = 1366
	picHeight      = 1024
	scale          = 7.0
	stylesName     = "style.yml"
	defFillColor   = "#757575AA"
	defBorderColor = "#0f0f0fAA"
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
	dc.InvertY()

	for _, feature := range fc.Features {

		fmt.Printf("\nType: %s\n", feature.Geometry.Type)

		switch feature.Geometry.Type {
		case "Polygon":
			for _, geom := range feature.Geometry.Polygon {
				firstPoint, x0, y0 := true, 0.0, 0.0
				for _, point := range geom {
					x := point[0]*scale + correct
					y := point[1]*scale + correct
					if firstPoint {
						x0 = x
						y0 = y
						firstPoint = false
					}
					dc.LineTo(x, y)
				}
				dc.LineTo(x0, y0)
				setStyle(dc, styles, feature.Properties)
				//level, _ := strconv.Atoi(feature.Properties["admin_level"].(string))
				//drawPoly(dc)
			}
		case "LineString":
			for _, geom := range feature.Geometry.LineString {
				x := geom[0] + correct
				y := geom[1] + correct
				dc.LineTo(x, y)
			}
			setLineStyle(dc, styles, feature.Properties)
			//drawLine(dc)
		case "Point":
			x := feature.Geometry.Point[0] + correct
			y := feature.Geometry.Point[1] + correct
			dc.DrawPoint(x, y, pointR)
			setStyle(dc, styles, feature.Properties)
			//drawP(dc)
		case "MultiPolygon":
			//level, _ := strconv.Atoi(feature.Properties["admin_level"].(string))
			for _, geom := range feature.Geometry.MultiPolygon {
				for _, poly := range geom {
					firstPoint, x0, y0 := true, 0.0, 0.0
					for _, point := range poly {
						x := 0.0
						if point[0] < -167.0 {
							x = (point[0]+360)*scale + correct
						} else {
							x = point[0]*scale + correct
						}
						y := point[1]*scale + correct
						if firstPoint {
							dc.MoveTo(x, y)
							x0 = x
							y0 = y
							firstPoint = false
						} else {
							dc.LineTo(x, y)
						}
					}
					dc.LineTo(x0, y0)
				}
				//drawPoly(dc, styles.AdLevel[level].FillColor, styles.AdLevel[level].BorderColor, styles.AdLevel[level].LineWidth)
				setStyle(dc, styles, feature.Properties)
			}
		}
	}

	dc.SavePNG("out.png")
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
