package main

// hull -- draw a limit line using the calculation for a
// lower hull. It has odd properties.

import (
	"encoding/csv"
	"flag"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

// Point is an x-y pair
type Point struct {
	X, Y float64
}

// main interprets the options and args.
func main() {
	var verbose bool
	var err error
	var minX, maxY float64

	flag.Float64Var(&minX, "minX", 0, "Set minimum x-value")
	flag.Float64Var(&maxY, "maxY", math.MaxFloat64, "Set maximum y-value")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime) // show file:line in logs

	if flag.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: hull [-v] load-file.csv\n") //nolint
		flag.PrintDefaults()
		os.Exit(1)
	}
	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("No load-test csv file provided, halting.\n")
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening %s: %s, halting.", filename, err)
	}
	defer f.Close() // nolint

	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	points := readCsv(r, filename, minX, maxY, verbose)
	start, end, _, _ := FindLowerHullLine(points, minX, maxY, verbose)
	m, b := slopeIntercept(start.X, start.Y, end.X, end.Y)
	// write just m and b to stdout
	fmt.Printf("%vx %v\n", m, b)
	log.Printf("In line (%v,%v) to (%v,%v)\n\t"+
		"the x-intercept is (%v,0)\n\t"+
		"and the equation is y = mx + b = %vx + %v\n",
		start.X, start.Y, end.X, end.Y, -b/m, m, b)
	plotPointsAndLine(points, start, end, minX, maxY, "lower_hull.png")

}

// FindLowerHullLine Finds the lowest point as the starting point,
// searches for the best endpoint that ensures no other points are
// below the line and returns the start and end points of the hull line
func FindLowerHullLine(points []Point, minX, maxY float64, verbose bool) (Point, Point, float64, float64) {
	low := 0.0
	high := 400.0

	if len(points) < 2 {
		return Point{}, Point{}, low, high
	}

	// Find highest right point
	start := points[0]
	for _, p := range points {
		if p.X > start.X {
			start = p
		}
	}
	if verbose {
		log.Printf("start point = %v\n", start)
	}

	// Find the best endpoint to the left of the start
	var bestEnd Point
	found := false

	// check if no other points fall below the potential line
	// This a uses a cross-product calculation and it's O(N^2). See below.
	for _, candidate := range points {
		// ignore points to the right of start
		if candidate.X >= start.X {
			continue
		}
		// check bounds
		if candidate.X < minX || candidate.Y > maxY {
			continue
		}

		// Nested duplicate loop, making it O(n^2)
		valid := true
		for _, p := range points {
			// ignore points to the right
			if p.X >= start.X || p == candidate {
				continue
			}

			// Check if point is below the line
			//if isPointBelowLine(start, candidate, p) {
			// Check if point is above the line
			if !isPointBelowLine(start, candidate, p) {
				valid = false
				break
			}
		}

		if valid && (!found || candidate.X < bestEnd.X) {
			bestEnd = candidate
			found = true
		}
	}
	if verbose {
		log.Printf("best end-point = %v\n", bestEnd)
	}

	return start, bestEnd, low, high
}

// isPointBelowLine uses a cross-product to see
// if the point is below line
func isPointBelowLine(start, end, point Point) bool {
	return (end.X-start.X)*(point.Y-start.Y)-
		(end.Y-start.Y)*(point.X-start.X) < 0
}

// plotPointsAndLine does just that
func plotPointsAndLine(points []Point, start, end Point, low, high float64, filename string) {
	p := plot.New()
	p.Title.Text = "Right Hull-Line"
	p.X.Label.Text = "Load, Requests per Second"
	p.Y.Label.Text = "Response Time, Seconds"

	p.X.Min = low
	p.X.Max = high

	// Plot points
	pts := make(plotter.XYs, len(points))
	for i, point := range points {
		pts[i].X = point.X
		pts[i].Y = point.Y
	}

	scatter, _ := plotter.NewScatter(pts)
	scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 0, A: 255}
	scatter.GlyphStyle.Radius = vg.Points(3)
	p.Add(scatter)

	// Plot line
	line := plotter.XYs{
		{X: start.X, Y: start.Y},
		{X: end.X, Y: end.Y},
	}

	linePlot, _ := plotter.NewLine(line)
	linePlot.Color = color.RGBA{G: 255, A: 255}
	linePlot.Width = vg.Points(2)
	p.Add(linePlot)

	// Save the plot
	p.Save(12*vg.Inch, 6*vg.Inch, filename)
}

// slope intercept generates a t = mx + b equation form
// a pair of points
func slopeIntercept(x1, y1, x2, y2 float64) (float64, float64) {
	m := (y2 - y1) / (x2 - x1)
	b := y2 - (x2 * m)
	return m, b
}

// readCsv reads preselected latency and requests per second from a csv file.
// that's the "perf" format, like "2018-01-17 10:40:38 0.00374 0.000185 0 5151 8"
func readCsv(r *csv.Reader, filename string, minX, maxY float64, verbose bool) []Point {
	const (
		latency = 2
		TPS     = 6
	)
	var recNo = 0
	var point Point
	var points = make([]Point, 0)

	if verbose {
		log.Printf("minX = %v, maxY = %v\n", minX, maxY)
	}
forloop:
	for ; ; recNo++ {
		record, err := r.Read()
		if verbose {
			log.Printf("read record = %q\n", record)
		}
		switch {
		case err == io.EOF:
			break forloop
		case err != nil:
			log.Fatalf("Fatal error mid-way reading %q from %s, stopping: %s\n", record, filename, err)
			break forloop
		}
		if len(record) < 7 {
			log.Printf("ill-formed record %q ignored\n",
				record)
			// Warning: this discards partial records
			continue
		}

		x, err := strconv.ParseFloat(record[TPS], 64)
		if err != nil {
			log.Fatalf("x in line %d of %q is invalid: %s\n", recNo, filename, record[TPS])
		}
		y, err := strconv.ParseFloat(record[latency], 64)
		if err != nil {
			log.Fatalf("y in line %d of %q is invalid: %s\n", recNo, filename, record[TPS])
		}

		// prune off values below minX or above Ymax
		if x < minX || y > maxY {
			continue
		}

		// create a new point to add
		point.X = x
		point.Y = y
		points = append(points, point)
	}
	return points
}
