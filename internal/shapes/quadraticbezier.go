package shapes

import "math"

type QuadraticBezier struct {
	Start   Point
	Control Point
	End     Point
}

func (qb *QuadraticBezier) Move(x, y float64) {
	qb.Start.Move(x, y)
	qb.Control.Move(x, y)
	qb.End.Move(x, y)
}

func (qb *QuadraticBezier) Transform(a, b, c, d, e, f float64) {
	qb.Start.Transform(a, b, c, d, e, f)
	qb.Control.Transform(a, b, c, d, e, f)
	qb.End.Transform(a, b, c, d, e, f)
}

func (qb *QuadraticBezier) Scale(factor float64, centroid Point) {
	qb.Start.Scale(factor, centroid)
	qb.Control.Scale(factor, centroid)
	qb.End.Scale(factor, centroid)
}

func (qb *QuadraticBezier) Rotate(angle float64, origin Point) {
	qb.Start.Rotate(angle, origin)
	qb.Control.Rotate(angle, origin)
	qb.End.Rotate(angle, origin)
}

func (cb *QuadraticBezier) PixelToMM(dpi float64) {
	cb.Start.PixelToMM(dpi)
	cb.Control.PixelToMM(dpi)
	cb.End.PixelToMM(dpi)
}

func (qb *QuadraticBezier) ToPolyline(bezierTolerance float64) *Polyline {
	const minSegments = 1
	const maxSegments = 512

	start := qb.Start
	control := qb.Control
	end := qb.End

	scale := maxQuadraticBezierScale(start, control, end)
	if scale == 0 {
		points := []Point{start, end}
		return NewPolyline(&points, nil)
	}

	toleranceSq := math.Pow(scale*bezierTolerance, 2)
	segments := minSegments
	var points []Point

	for {
		points = make([]Point, segments+1)
		invSegments := 1.0 / float64(segments)
		for i := 0; i <= segments; i++ {
			t := float64(i) * invSegments
			points[i] = quadraticBezierPoint(start, control, end, t)
		}

		if segments >= maxSegments || quadraticBezierHausdorffSq(start, control, end, points) <= toleranceSq {
			break
		}

		segments *= 2
	}

	return NewPolyline(&points, nil)
}

func quadraticBezierPoint(start, control, end Point, t float64) Point {
	mt := 1 - t

	ax := mt*start.X + t*control.X
	ay := mt*start.Y + t*control.Y
	bx := mt*control.X + t*end.X
	by := mt*control.Y + t*end.Y

	return Point{
		X: mt*ax + t*bx,
		Y: mt*ay + t*by,
	}
}

func quadraticBezierHausdorffSq(start, control, end Point, points []Point) float64 {
	if len(points) < 2 {
		return 0
	}

	maxDistanceSq := 0.0
	segments := len(points) - 1
	invSegments := 1.0 / float64(segments)
	for i := 0; i < segments; i++ {
		tMid := (float64(i) + 0.5) * invSegments
		midPoint := quadraticBezierPoint(start, control, end, tMid)
		distanceSq := pointToSegmentDistanceSq(midPoint, points[i], points[i+1])
		if distanceSq > maxDistanceSq {
			maxDistanceSq = distanceSq
		}
	}

	return maxDistanceSq
}

func pointToSegmentDistanceSq(p, a, b Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return (p.X-a.X)*(p.X-a.X) + (p.Y-a.Y)*(p.Y-a.Y)
	}

	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	px := a.X + t*dx
	py := a.Y + t*dy
	dx = p.X - px
	dy = p.Y - py
	return dx*dx + dy*dy
}

func maxQuadraticBezierScale(start, control, end Point) float64 {
	maxDistance := start.DistanceTo(&control)
	distance := control.DistanceTo(&end)
	if distance > maxDistance {
		maxDistance = distance
	}
	distance = start.DistanceTo(&end)
	if distance > maxDistance {
		maxDistance = distance
	}
	return maxDistance
}
