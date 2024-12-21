# Lower hull, AKA bounds line

courtesy of Claude 3.5 Sonnet
for David Collier-Brown

## Q1
with pencil and straightedge, draw a line on a cloud of points,  between the lowest point and one of the points to its right, such that no point in between is below the line


## A1
Let me explain this geometric algorithm clearly:

This is known as the "Lower Hull" problem and can be solved efficiently. Here's how to draw the line:

Find the leftmost point with the lowest y-coordinate
For each point to the right of it, calculate the slope between the lowest point and that point
Choose the point that creates a line where no other points fall below it
The key insight is that this forms part of a convex hull construction, specifically the lower portion. The line you draw will be part of the lower hull of the point set.

A practical way to find this line:

Start at the lowest point
Try connecting it to each point to its right
Check if any points lie below each candidate line
Select the line that keeps all points above or on it
This is similar to the first step in Graham's scan algorithm for convex hull construction, but focused only on finding a valid lower hull segment from the leftmost lowest point.

The resulting line will be a supporting line of the point set, guaranteeing no points fall below it.

## Q2 
describe how to do this in golang

## A2
Here's a clean implementation in Go that solves the lower hull line problem:

(skip)



The code:

Finds the lowest point (leftmost if tied)
Examines all points to the right
Uses cross product to check if any points fall below candidate lines
Returns the valid line segment with the leftmost endpoint
The algorithm runs in O(n²) time where n is the number of points. The cross product method provides a robust way to determine if points are below a line without dealing with floating-point slope calculations.

To test it, you can create a slice of Points and pass them to FindLowerHullLine, which returns the two Points that define the line segment.

## Q3
add plotting the line to the above

## A4
Let's add plotting using the popular "gonum/plot" package to visualize the points and the line:

(skip)

To run this, first install the required package:

go get gonum.org/v1/plot

This enhanced version:

Creates a new plot
Plots all points as red dots
Draws the lower hull line in green
Saves the result as a PNG file
The output will be saved as "lower_hull.png" in your working directory, showing:

All points as red dots
The computed lower hull line in green
Labeled axes and title
The visualization makes it clear that no points fall below the computed line.

## Q5
where can I read about lower hulls. assume I'm dumb

## A5
You're asking a great question! Let's start with some beginner-friendly resources:

Start with this visual explanation of convex hulls (lower hulls are part of this): https://www.algorithmtutor.com/computational-geometry/convex-hull-algorithms/

For a gentle introduction with pictures:

"Computational Geometry in C" by Joseph O'Rourke - Chapter 3 has excellent diagrams
GeeksForGeeks has a clear walkthrough: https://www.geeksforgeeks.org/convex-hull-set-1-jarviss-algorithm-or-wrapping/
For hands-on learning:

Try the interactive visualizer at https://visualgo.net/en/convexhull
Draw points on paper and practice finding lower hulls by hand
The key concepts build step by step:

First understand what "below a line" means
Then learn about point-line relationships
Finally see how lower hulls fit into convex hulls
Everyone starts somewhere! These resources are designed to build understanding from the ground up.


