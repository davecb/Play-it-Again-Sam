package main

import (
    "fmt"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
    "image/color"
)

func main() {
    points := []Point{
        {0, 0},
        {1, 1},
        {2, 0.5},
        {3, 2},
        {4, 0.99},
    }
    
    start, end := FindLowerHullLine(points)
    fmt.Printf("Line from (%v,%v) to (%v,%v)\n", 
        start.X, start.Y, end.X, end.Y)
    plotPointsAndLine(points, start, end, "lower_hull.png")

}

type Point struct {
    X, Y float64
}

func FindLowerHullLine(points []Point) (Point, Point) {
    if len(points) < 2 {
        return Point{}, Point{}
    }

    // Find lowest point
    start := points[0]
    for _, p := range points {
        if p.Y < start.Y || (p.Y == start.Y && p.X < start.X) {
            start = p
        }
    }

    // Find best endpoint to the right
    var bestEnd Point
    found := false

    for _, candidate := range points {
        if candidate.X <= start.X {
            continue
        }

        valid := true
        for _, p := range points {
            if p.X <= start.X || p == candidate {
                continue
            }

            // Check if point is below the line
            if isPointBelowLine(start, candidate, p) {
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

func isPointBelowLine(start, end, point Point) bool {
    // Cross product to determine if point is below line
    return (end.X-start.X)*(point.Y-start.Y) - 
           (end.Y-start.Y)*(point.X-start.X) < 0
}


func plotPointsAndLine(points []Point, start, end Point, filename string) {
    p := plot.New()
    p.Title.Text = "Lower Hull Line"
    p.X.Label.Text = "X"
    p.Y.Label.Text = "Y"

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

