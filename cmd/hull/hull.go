package main

// hull -- draw a limit line using the calculation for a
// lower hull. It has odd properties.

import (
    "fmt"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
    "image/color"
    "github.com/davecb/Play-it-Again-Sam/cmd/hull/decl"
    "github.com/davecb/Play-it-Again-Sam/cmd/hull/data"
)

func main() {
	points := data.Get()
//     points := []Point{
//         {0, 0.6},
////         {1, 1},
////         {2, 0.5},
////         {3, 2},
////         {4, 0.99},
////     }
//
//       points := []Point{ 
////	{ 252, 0.64 },
////	{ 252, 1.72 },
////	{ 253, 6.18 },
////	{ 254, 0.78 },
////	{ 255, 2.51 },
////	{ 255, 0.50 },
////	{ 256, 0.54 },
////	{ 256, 4.75 },
////	{ 256, 1.34 },
////	{ 257, 0.99 },
////	{ 258, 0.99 },
////	{ 261, 3.26 },
////	{ 262, 59.63 },
////	{ 262, 8.43 },
////	{ 263, 0.97 },
////	{ 265, 4.12 },
////	{ 266, 0.48 },
////	{ 267, 2.33 },
////	{ 269, 2.37 },
////	{ 269, 51.39 },
////	{ 270, 1.87 },
////	{ 270, 5.41 },
////	{ 271, 0.83 },
////	{ 271, 3.11 },
////	{ 272, 1.43 },
////	{ 272, 11.60 },
////	{ 273, 1.99 },
////	{ 273, 2.36 },
////	{ 273, 5.73 },
////	{ 273, 51.07 },
////	{ 274, 8.41 },
////	{ 275, 6.81 },
////	{ 276, 5.80 },
////	{ 279, 93.75 },
////	{ 281, 5.32 },
////	{ 284, 56.73 },
////	{ 285, 43.95 },
////	{ 290, 11.15 },
////	{ 292, 21.76 },
////	{ 292, 46.30 },
////	{ 293, 10.81 },
////	{ 294, 8.70 },
////	{ 294, 10.99 },
////	{ 296, 22.83 },
////	{ 298, 11.89 },
////	{ 299, 15.12 },
////	{ 299, 2.31 },
//	{ 301, 1.52 },
//	{ 302, 24.39 },
//	{ 302, 16.93 },
//	{ 302, 17.15 },
//	{ 302, 78.64 },
//	{ 303, 16.64 },
//	{ 304, 14.62 },
//	{ 304, 16.17 },
//	{ 304, 2.63 },
//	{ 304, 4.99 },
//	{ 305, 13.38 },
//	{ 305, 17.10 },
//	{ 306, 18.03 },
//	{ 307, 14.62 },
//	{ 308, 21.81 },
//	{ 308, 20.32 },
//	{ 308, 51.35 },
//	{ 308, 58.45 },
//	{ 310, 19.80 },
//	{ 310, 20.90 },
// as used for the last experiments
//	{ 312, 21.53 },
//	{ 313, 18.82 },
//	{ 314, 18.67 },
//	{ 315, 5.13 },
//	{ 316, 19.08 },
//	{ 316, 21.12 },
//	{ 320, 72.47 },
//	{ 321, 21.65 },
//	{ 322, 21.94 },
//	{ 323, 25.35 },
//	{ 324, 25.96 },
//	{ 328, 24.86 },
//	{ 328, 71.87 },
//	{ 328, 33.76 },
//	{ 328, 33.68 },
//	{ 330, 35.13 },
//	{ 334, 41.94 },
//	{ 338, 50.90 },
//	{ 339, 51.61 },
//}
    
    start, end := FindLowerHullLine(points)
    fmt.Printf("Line from (%v,%v) to (%v,%v)\n", 
        start.X, start.Y, end.X, end.Y)
    plotPointsAndLine(points, start, end, "lower_hull.png")

}
//
//type Point struct {
//    X, Y float64
//}

// FindLowerHullLine Finds the lowest point as the starting point,
// searches for the best endpoint that ensures no other points are 
// below the line and returns the start and end points of the hull line
func FindLowerHullLine(points []decl.Point) (decl.Point, decl.Point) {
    if len(points) < 2 {
        return decl.Point{}, decl.Point{}
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
    var bestEnd decl.Point
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
func isPointBelowLine(start, end, point decl.Point) bool {
    return (end.X-start.X)*(point.Y-start.Y) - 
           (end.Y-start.Y)*(point.X-start.X) < 0
}


// plotPointsAndLine does just that
func plotPointsAndLine(points []decl.Point, start, end decl.Point, filename string) {
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

