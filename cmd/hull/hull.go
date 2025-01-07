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
	"os"
	"strconv"
	// "github.com/davecb/Play-it-Again-Sam/cmd/hull/decl"
)

// Point is an x-y pair
type Point struct {
	X, Y float64
}

// main interprets the options and args.
func main() {
	var verbose bool
	var err error

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

	points := readCsv(r, filename, verbose)
	start, end := FindLowerHullLine(points)
	fmt.Printf("Line from (%v,%v) to (%v,%v)\n",
		start.X, start.Y, end.X, end.Y)
	plotPointsAndLine(points, start, end, "lower_hull.png")

}

// FindLowerHullLine Finds the lowest point as the starting point,
// searches for the best endpoint that ensures no other points are
// below the line and returns the start and end points of the hull line
func FindLowerHullLine(points []Point) (Point, Point) {
	if len(points) < 2 {
		return Point{}, Point{}
	}

	// Find lowest point, preferring leftmost ones ...
	// changed to rightmost
	// changed to highest right
	start := points[0]
	for _, p := range points {
		//if p.Y < start.Y || (p.Y == start.Y && p.X < start.X) {
		//    start = p
		//}
		if p.X > start.X {
			start = p
		}
	}

	// Find best endpoint to the right
	var bestEnd Point
	found := false

	// check if no other points fall below the potential line
	// This a uses a cross-product calculation and it's O(N^2)
	for _, candidate := range points {
		// ignore points to the left of start
		// if candidate.X <= start.X {
		// ignore points to the right of start
		if candidate.X >= start.X {
			continue
		}

		// Duplicate inner loop, making it O(n^2)
		valid := true
		for _, p := range points {
			//// ignore points to the left
			//if p.X <= start.X || p == candidate {
			//    continue
			// ignore points to the left
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

	return start, bestEnd
}

// isPointBelowLine uses a cross-product to see
// if the point is below line
func isPointBelowLine(start, end, point Point) bool {
	return (end.X-start.X)*(point.Y-start.Y)-
		(end.Y-start.Y)*(point.X-start.X) < 0
}

// plotPointsAndLine does just that
func plotPointsAndLine(points []Point, start, end Point, filename string) {
	p := plot.New()
	p.Title.Text = "Right Hull-Line"
	p.X.Label.Text = "Load, Requests per Second"
	p.Y.Label.Text = "Response Time, Seconds"

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
	p.Save(6*vg.Inch, 6*vg.Inch, filename)
}

// readCsv reads preselected latency and requests per second from a csv file.
// that's the "perf" format, like "2018-01-17 10:40:38 0.00374 0.000185 0 5151 8"
func readCsv(r *csv.Reader, filename string, verbose bool) []Point {
	const (
		latency = 7
		TPS     = 3
	)
	var recNo = 0
	var point Point
	var points = make([]Point, 0)
forloop:
	for {
		record, err := r.Read()
		switch {
		case err == io.EOF:
			if verbose {
				log.Printf("At EOF on %s\n", filename)
			}
			break forloop
		case err != nil:
			log.Printf("Fatal error mid-way reading %q from %s, stopping: %s\n", record, filename, err)
			break forloop
		}
		if len(record) != 7 {
			log.Printf("ill-formed record %q ignored\n",
				record)
			// Warning: this discards real-time part-records
			continue
		}

		x, err := strconv.ParseFloat(record[TPS], 64)
		if err != nil {
			log.Fatalf("x in line %d of %q is invalid: %s\n", recNo, filename, x)
		}
		y, err := strconv.ParseFloat(record[latency], 64)
		if err != nil {
			log.Fatalf("y in line %d of %q is invalid: %s\n", recNo, filename, x)
		}
		point.X = x
		point.Y = y
		points = append(points, point)
	}
	return nil
}
