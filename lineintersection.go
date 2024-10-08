package main

// Determines if the line segments p1---p2 & p3---p4 intersect
func intersectingLineSegments(p1, p2, p3, p4 *Point) bool {
	if pointOnLine(p1, p2, p3) || pointOnLine(p1, p2, p4) {
		return true
	}

	return (counterClockwise(p1, p3, p4) != counterClockwise(p2, p3, p4)) &&
		(counterClockwise(p1, p2, p3) != counterClockwise(p1, p2, p4))
}

// Determines the point where p1---p2 & p3---p4 intersect
func calculateLineIntersection(p1, p2, p3, p4 *Point) (Point, bool) {
	denom := (p4.Y-p3.Y)*(p2.X-p1.X) - (p4.X-p3.X)*(p2.Y-p1.Y)
	if denom == 0 {
		return Point{0, 0}, false
	}
	ua := ((p4.X-p3.X)*(p1.Y-p3.Y) - (p4.Y-p3.Y)*(p1.X-p3.X)) / denom
	if ua < 0 || ua > 1 {
		return Point{0, 0}, false
	}
	ub := ((p2.X-p1.X)*(p1.Y-p3.Y) - (p2.Y-p1.Y)*(p1.X-p3.X)) / denom
	if ub < 0 || ub > 1 {
		return Point{0, 0}, false
	}

	return Point{p1.X + ua*(p2.X-p1.X), p1.Y + ua*(p2.Y-p1.Y)}, true
}

func intersectingPolylines(line1, line2 *Polyline) bool {
	for i := 1; i < len(line1.points); i++ {
		for j := 1; j < len(line2.points); j++ {
			if intersectingLineSegments(&line1.points[i-1], &line1.points[i], &line2.points[j-1], &line2.points[j]) {
				return true
			}
		}
	}
	return false
}

func lineIntersectsLineSegment(line *Polyline, p1, p2 *Point) bool {
	for i := 1; i < len(line.points); i++ {
		if intersectingLineSegments(&line.points[i-1], &line.points[i], p1, p2) {
			return true
		}
	}

	return false
}

//Counts the how many line segments of line1 intersect line2
func countLineIntersections(line1, line2 *Polyline) int {
	intersections := 0
	for i := 1; i < len(line1.points); i++ {
		if lineIntersectsLineSegment(line2, &line1.points[i-1], &line1.points[i]) {
			intersections++
		}
	}
	return intersections
}

// Determines if the point p3 is on the line p1---p2
func pointOnLine(p1, p2, p3 *Point) bool {
	return p1.distanceTo(p3)+p2.distanceTo(p3) == p1.distanceTo(p2)
}

func counterClockwise(p1 *Point, p2 *Point, p3 *Point) bool {
	return (p3.Y-p1.Y)*(p2.X-p1.X) > (p2.Y-p1.Y)*(p3.X-p1.X)
}
