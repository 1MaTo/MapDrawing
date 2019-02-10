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
	correct = 500
	pointR  = 1
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

	dc := gg.NewContext(1366, 1024)
	dc.InvertY()

	for _, feature := range fc.Features {

		fmt.Printf("\nType: %s\n", feature.Geometry.Type)

		switch feature.Geometry.Type {
		case "Polygon":
			for _, geom := range feature.Geometry.Polygon {
				for _, point := range geom {
					x := point[0] + correct
					y := point[1] + correct
					dc.LineTo(x, y)
				}
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
