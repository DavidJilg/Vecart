package shapes

import "math"

type CubicBezier struct {
	Start    Point
	Control1 Point
	Control2 Point
	End      Point
}

func (cb *CubicBezier) Move(x, y float64) {
	cb.Start.Move(x, y)
	cb.Control1.Move(x, y)
	cb.Control2.Move(x, y)
	cb.End.Move(x, y)
}

func (cb *CubicBezier) Transform(a, b, c, d, e, f float64) {
	cb.Start.Transform(a, b, c, d, e, f)
	cb.Control1.Transform(a, b, c, d, e, f)
	cb.Control2.Transform(a, b, c, d, e, f)
	cb.End.Transform(a, b, c, d, e, f)
}

func (cb *CubicBezier) Scale(factor float64, centroid Point) {
	cb.Start.Scale(factor, centroid)
	cb.Control1.Scale(factor, centroid)
	cb.Control2.Scale(factor, centroid)
	cb.End.Scale(factor, centroid)
}

func (cb *CubicBezier) Rotate(angle float64, origin Point) {
	cb.Start.Rotate(angle, origin)
	cb.Control1.Rotate(angle, origin)
	cb.Control2.Rotate(angle, origin)
	cb.End.Rotate(angle, origin)
}

func (cb *CubicBezier) PixelToMM(dpi float64) {
	cb.Start.PixelToMM(dpi)
	cb.End.PixelToMM(dpi)
	cb.Control1.PixelToMM(dpi)
	cb.Control2.PixelToMM(dpi)
}

func (cb *CubicBezier) ToPolyline(bezierTolerance float64) *Polyline {
	const minSegments = 1
	const maxSegments = 512

	start := cb.Start
	control1 := cb.Control1
	control2 := cb.Control2
	end := cb.End

	scale := maxCubicBezierScale(start, control1, control2, end)
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
			points[i] = cubicBezierPoint(start, control1, control2, end, t)
		}

		if segments >= maxSegments || cubicBezierHausdorffSq(start, control1, control2, end, points) <= toleranceSq {
			break
		}

		segments *= 2
	}

	return NewPolyline(&points, nil)
}

func cubicBezierPoint(start, control1, control2, end Point, t float64) Point {
	mt := 1 - t

	ax := mt*start.X + t*control1.X
	ay := mt*start.Y + t*control1.Y
	bx := mt*control1.X + t*control2.X
	by := mt*control1.Y + t*control2.Y
	cx := mt*control2.X + t*end.X
	cy := mt*control2.Y + t*end.Y

	dx := mt*ax + t*bx
	dy := mt*ay + t*by
	ex := mt*bx + t*cx
	ey := mt*by + t*cy

	return Point{
		X: mt*dx + t*ex,
		Y: mt*dy + t*ey,
	}
}

func cubicBezierHausdorffSq(start, control1, control2, end Point, points []Point) float64 {
	if len(points) < 2 {
		return 0
	}

	maxDistanceSq := 0.0
	segments := len(points) - 1
	invSegments := 1.0 / float64(segments)
	for i := 0; i < segments; i++ {
		tMid := (float64(i) + 0.5) * invSegments
		midPoint := cubicBezierPoint(start, control1, control2, end, tMid)
		distanceSq := pointToSegmentDistanceSq(midPoint, points[i], points[i+1])
		if distanceSq > maxDistanceSq {
			maxDistanceSq = distanceSq
		}
	}

	return maxDistanceSq
}

func maxCubicBezierScale(start, control1, control2, end Point) float64 {
	maxDistance := start.DistanceTo(&control1)
	distance := control1.DistanceTo(&control2)
	if distance > maxDistance {
		maxDistance = distance
	}
	distance = control2.DistanceTo(&end)
	if distance > maxDistance {
		maxDistance = distance
	}
	distance = start.DistanceTo(&end)
	if distance > maxDistance {
		maxDistance = distance
	}
	return maxDistance
}
