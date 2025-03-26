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
	"sort"
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
	// sort from high to low x-values
	sort.Slice(points, func(i, j int) bool {
		return points[i].X > points[j].X
	})
	err, start, end := FindLowerHullLine(points, minX, maxX, maxY, verbose)
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

	plotPointsAndLine(points, start, end, Point{-b / m, 0.0}, "lower_hull.png")
	log.Printf("A graph of the data is in lower_hull.png")
}

// FindLowerHullLine Finds the lowest point as the starting point,
// searches for the best endpoint that ensures no other points are
// below the line and returns the start and end points of the hull line
func FindLowerHullLine(points []Point, minX, maxX, maxY float64, verbose bool) (error, Point, Point) {
	var bestEnd Point

	if l := len(points); l < 2 {
		return fmt.Errorf("we require at least 3 points, got %d", l), Point{}, Point{}
	}
	points = trimPoints(points, minX, maxX, maxY)

	// Find rightmost point to use as the start
	// This can require use of maX or maxY to skip outliers!
	start := points[0]
	for _, p := range points {
		if p.X > start.X {
			start = p
		}
	}
	log.Printf("%f,%f\n", start.X, start.Y)
	// postcondition: p is the rightmost point, bestEnd is uninitialized (0,0) FIXME

	// Loop and find the most-right endpoint below and to the left of the start point

	// check if no other points fall below the potential line
	// This re-iterates across all the points, so it's therefor O(N^2).
	found := false // FIXME ?
	for _, candidate := range points {
		log.Printf("   %f,%f\n", candidate.X, candidate.Y)
		// ignore points to the right of start (which should be the rightmost)
		//if candidate.X > start.X {
		//	continue // this should never happen
		//}

		// iterate across all the points, looking for ones further below the line
		// Nested duplicate loop, making it O(n^2)
		valid := true
		for _, innerP := range points {
			//Ignore self and points to the right of start
			// FIXME, copied back in
			if innerP.X >= start.X || innerP == candidate {
				continue
			}

			// Check if point is below and right of the previous candidate
			if !isPointBelowLine(start, candidate, innerP) {
				// this point is *not* below and right of the candidate, ignore it and try another
				//continue
				valid = false
				//break // FIXME, what am I doing? Exit the inner loop on the first non-below case???
				// maybe just mark invalid
			}
			// postcondition: we are only looking at points below and to the right of the line
			// from start to the previous candidate. This is a possible candidate

			// try to see if it has a lower X than bestEnd
			if valid && (!found || candidate.X < bestEnd.X) {
				//if !found || candidate.X < bestEnd.X
				//if candidate.X > bestEnd.X
				bestEnd = candidate
				found = true
				log.Printf("   maybe best: %f,%f\n", candidate.X, candidate.Y)
			}
		}
		// loop postcondition: we have the lowest rightmost point
	}
	log.Printf("best end-point = %v\n", bestEnd)

	return nil, start, bestEnd
}

// trimPoints removes ones outside the specified rangs
func trimPoints(points []Point, minX float64, maxX float64, maxY float64) []Point {
	var hullPoints []Point

	// create a smaller map using minX, etc
	for _, p := range points {
		// discard specified points
		//log.Printf("trim: point Y = %f, maxY =  %f\n", p.Y, maxY)
		if p.X < minX {
			log.Printf("skipped, X < minX\n")
			continue
		}
		if p.X > maxX {
			log.Printf("skipped, X > maxX\n")
			continue
		}
		if p.Y > maxY {
			// log.Printf("skipped, Y > maxY\n")
			continue
		}
		hullPoints = append(hullPoints, p)
	}
	return hullPoints
}

// isPointBelowLine uses a cross-product to see
// if the point is below line
func isPointBelowLine(lineStart, lineEnd, candidate Point) bool {
	q := ((lineEnd.X-lineStart.X)*(candidate.Y-lineStart.Y)-
		(lineEnd.Y-lineStart.Y)*(candidate.X-lineStart.X) < 0)
	if q {
		log.Printf("returning true, %v\n", candidate)
	}
	return q
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
			break forloop
		}
		if len(record) < 7 {
			log.Printf("ill-formed record %q ignored\n",
				record)
			// Warning: this discards partial records
			continue
		}

		x, err := strconv.ParseFloat(record[TPS], 64) // requests, x
		if err != nil {
			log.Fatalf("x in line %d of %q is invalid: %s\n", recNo, filename, record[TPS])
		}
		y, err := strconv.ParseFloat(record[latency], 64) // response time, y
		if err != nil {
			log.Fatalf("y in line %d of %q is invalid: %s\n", recNo, filename, record[TPS])
		}

		// create a new point to add
		point.X = x
		point.Y = y
		points = append(points, point)
	}
	return points
}
