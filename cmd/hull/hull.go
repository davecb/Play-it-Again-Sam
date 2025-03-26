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
	//"sort"
	"strconv"
)

// Point is an x-y pair
type Point struct {
	X, Y float64
}

// main interprets the options and args.
func main() {
	var verbose bool
	var minX, maxX, maxY float64
	var err error

	//flag.Float64Var(&minX, "minX", 0, "Set minimum x-value")
	//flag.Float64Var(&maxY, "maxY", math.MaxFloat64, "Set maximum y-value")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Float64Var(&minX, "minX", 0, "ignore points below this minimum x-value")               // for convenience/speed
	flag.Float64Var(&maxX, "maxX", math.MaxFloat32, "ignore points above this maximum x-value") // for excluding points
	flag.Float64Var(&maxY, "maxY", math.MaxFloat32, "ignore points above this maximum y-value") // for excluding points

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

	rawPoints := readCsv(r, filename, verbose)
	points := trimPoints(rawPoints, minX, maxX, maxY)
	// sort from high to low x-values FIXME
	//sort.Slice(points, func(i, j int) bool {
	//	return points[i].X > points[j].X
	//})
	start, end, err := FindLowerHullLine(points, verbose)
	if err != nil {
		log.Fatalf("failure in FindLowerHullLine, %v", err)
	}
	m, b := slopeIntercept(start.X, start.Y, end.X, end.Y)
	// write m, b to and x-intercept to stdout
	fmt.Printf("%vx %v # y = mx+b\n", m, b)
	fmt.Printf("%v %v # start\n", start.X, start.Y)
	fmt.Printf("%v %v # end\n", end.X, end.Y)
	fmt.Printf("%v %v # x-intercept\n", -b/m, 0)

	// write a user-focused description to stderr
	log.Printf("In line (%v,%v) to (%v,%v)\n\t"+
		"the x-intercept is (%v,0)\n\t"+
		"and the equation is y = mx + b = %vx + %v\n",
		start.X, start.Y, end.X, end.Y, -b/m, m, b)

	//plotPointsAndLine(points, start, end, Point{-b / m, 0.0}, low, high,"lower_hull.png")
	plotPointsAndLine(points, start, end, Point{-b / m, 0.0}, "lower_hull.png")
	log.Printf("A graph of the data is in lower_hull.png")
}

// FindLowerHullLine Finds the lowest point as the starting point,
// searches for the best endpoint that ensures no other points are
// below the line and returns the start and end points of the hull line
func FindLowerHullLine(points []Point, verbose bool) (Point, Point, error) {
	var bestEnd Point

	if l := len(points); l < 3 {
		return Point{}, Point{}, fmt.Errorf("we require at least 3 points, got %d", l)
	}

	// Find rightmost point to use as the start
	// This can require use of maX or maxY to skip outliers!
	start := points[0]
	for _, p := range points {
		//if p.Y < start.Y || (p.Y == start.Y && p.X < start.X) {
		//    start = p
		//}
		if p.X > start.X {
			start = p
		}
	}
	// postcondition: p is the rightmost point, bestEnd is uninitialized (0,0)

	// Find the best endpoint to the left of the start, by
	// check if no other points fall below the potential line
	// This uses a cross-product calculation and it's O(N^2)
	found := false
	for _, candidate := range points {
		// ignore points to the right of start
		if candidate.X >= start.X {
			continue // Can't happen (:-))
		}

		// iterate across all the points, looking for ones below the line
		valid := true
		for _, p := range points {
			// ignore points to the right
			if p.X >= start.X || p == candidate {
				continue
			}
			if p.X < .1 {
				// throw away points close to zero
				// this is a heuristic, and depends on the data
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
		// loop postcondition: we have the lowest rightmost point
	}
	if verbose {
		log.Printf("best end-point = %v\n", bestEnd)
	}
	return start, bestEnd, nil
}

// trimPoints removes ones outside the specified ranges
func trimPoints(points []Point, minX float64, maxX float64, maxY float64) []Point {
	var hullPoints []Point

	// create a smaller map using minX, etc
	for _, p := range points {
		// discard specified points
		if p.X < minX {
			continue
		}
		if p.X > maxX {
			continue
		}
		if p.Y > maxY {
			continue
		}
		hullPoints = append(hullPoints, p)
	}
	return hullPoints
}

// isPointBelowLine uses a cross-product to see
// if the point is below line
func isPointBelowLine(start, end, point Point) bool {
	return (end.X-start.X)*(point.Y-start.Y)-
		(end.Y-start.Y)*(point.X-start.X) < 0
}

// plotPointsAndLine does just that
func plotPointsAndLine(points []Point, lineStart, lineEnd, xIntercept Point, filename string) {
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
		{X: lineStart.X, Y: lineStart.Y},
		{X: lineEnd.X, Y: lineEnd.Y},
		{X: xIntercept.X, Y: xIntercept.Y},
	}

	linePlot, _ := plotter.NewLine(line)
	linePlot.Color = color.RGBA{G: 255, A: 255} // bright green
	linePlot.Width = vg.Points(2)
	p.Add(linePlot)

	// Save the plot
	p.Save(12*vg.Inch, 6*vg.Inch, filename)
}

// slope intercept generates a y = mx + b equation form
// a pair of points
func slopeIntercept(x1, y1, x2, y2 float64) (float64, float64) {
	m := (y2 - y1) / (x2 - x1)
	b := y2 - (x2 * m)
	return m, b
}

// readCsv reads preselected latency and requests per second from a csv file.
// that's the "perf" format, like "2018-01-17 10:40:38 0.00374 0.000185 0 5151 8"
func readCsv(r *csv.Reader, filename string, verbose bool) []Point {
	const (
		latency = 2
		TPS     = 6
	)
	var recNo = 0
	var point Point
	var points = make([]Point, 0)

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
		}
		if len(record) != 7 {
			log.Printf("ill-formed record %q ignored\n", record)
			// Warning: this intentionally discards partial records
			continue
		}

		x, err := strconv.ParseFloat(record[TPS], 64)
		if err != nil {
			log.Fatalf("x in line %d of %q is invalid: %s\n", recNo, filename, record[TPS])
		}
		y, err := strconv.ParseFloat(record[latency], 64) // response time, y
		if err != nil {
			log.Fatalf("y in line %d of %q is invalid: %s\n", recNo, filename, record[latency])
		}

		// create a new point to add
		point.X = x
		point.Y = y
		points = append(points, point)
	}
	return points
}
