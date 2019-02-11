package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/fogleman/gg"
	"github.com/paulmach/go.geojson"
)

var fileName *string

const (
	correct = 0
	pointR  = 1
	picWidth = 1366
	picHeight = 1024
	scale = 7.0
)

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

	dc := gg.NewContext(picWidth, picHeight)
	dc.InvertY()

	for _, feature := range fc.Features {

		fmt.Printf("\nType: %s\n", feature.Geometry.Type)

		switch feature.Geometry.Type {
		case "Polygon":
			for _, geom := range feature.Geometry.Polygon {
				firstPoint, x0, y0 := true, 0.0, 0.0
				for _, point := range geom {
					x := point[0] + correct
					y := point[1] + correct
					if (firstPoint){
						x0 = x
						y0 = y 
						firstPoint = false
					}
					dc.LineTo(x, y)
				}
				dc.LineTo(x0, y0)
				drawPoly(dc)
			}
		case "LineString":
			for _, geom := range feature.Geometry.LineString {
				x := geom[0] + correct
				y := geom[1] + correct
				dc.LineTo(x, y)

			}
			drawLine(dc)
		case "Point":
			x := feature.Geometry.Point[0] + correct
			y := feature.Geometry.Point[1] + correct
			dc.DrawPoint(x, y, pointR)
			drawP(dc)
		case "MultiPolygon":
			for _, geom := range feature.Geometry.MultiPolygon {
				for _, poly := range geom {
					firstPoint, x0, y0 := true, 0.0, 0.0
					for _, point := range poly{
						x := 0.0;
						if (point[0] < -167.0){
							x = (point[0]+360)  * scale  + correct
						}else{
							x = point[0] * scale + correct
						}
						y := point[1] * scale + correct
						if (firstPoint){
							dc.MoveTo(x, y)
							x0 = x
							y0 = y 
							firstPoint = false
						}else{
							dc.LineTo(x, y)
						}
					}
					dc.LineTo(x0, y0)
				}
				drawPoly(dc)
			}
		}
	}
	dc.SavePNG("out.png")
}

func drawPoly(dc *gg.Context) {
	dc.SetRGBA(0, 0, 0, 0.1)
	dc.FillPreserve()
	dc.SetRGBA(1, 0, 0, 0.5)
	dc.SetLineWidth(0.5)
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


