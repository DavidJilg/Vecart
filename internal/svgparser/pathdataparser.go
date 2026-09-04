package svgparser

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

func parsePathDataString(dString string) (shapes.Path, error) {
	dString = cleanString(dString)
	dString = strings.TrimSpace(dString)
	if dString == "" {
		return shapes.NewPath(nil), errors.New("empty path data string")
	}

	subpaths := make([][]any, 0, 8)
	currentSubpath := make([]any, 0, 64)
	var current shapes.Point
	var subpathStart shapes.Point
	var lastCubicControl *shapes.Point
	var lastQuadControl *shapes.Point
	var lastCmdUpper byte
	var cmd byte
	nums := make([]float64, 0, 16)

	flush := func() error {
		if cmd == 0 {
			return nil
		}
		upper := cmd
		if upper >= 'a' && upper <= 'z' {
			upper = upper - 'a' + 'A'
		}
		isRelative := cmd >= 'a' && cmd <= 'z'

		switch upper {
		case 'M':
			if len(nums) < 2 || len(nums)%2 != 0 {
				return errors.New("invalid moveto parameters")
			}
			for i := 0; i < len(nums); i += 2 {
				pt := shapes.Point{X: nums[i], Y: nums[i+1]}
				if isRelative {
					pt.X += current.X
					pt.Y += current.Y
				}
				if i == 0 {
					if len(currentSubpath) > 0 {
						subpaths = append(subpaths, currentSubpath)
						currentSubpath = make([]any, 0, 64)
					}
					current = pt
					subpathStart = pt
					currentSubpath = append(currentSubpath, &pt)
				} else {
					current = pt
					currentSubpath = append(currentSubpath, &pt)
				}
			}
			lastCubicControl = nil
			lastQuadControl = nil
		case 'L':
			if len(nums) < 2 || len(nums)%2 != 0 {
				return errors.New("invalid lineto parameters")
			}
			for i := 0; i < len(nums); i += 2 {
				pt := shapes.Point{X: nums[i], Y: nums[i+1]}
				if isRelative {
					pt.X += current.X
					pt.Y += current.Y
				}
				current = pt
				currentSubpath = append(currentSubpath, &pt)
			}
			lastCubicControl = nil
			lastQuadControl = nil
		case 'H':
			if len(nums) < 1 {
				return errors.New("invalid horizontal lineto parameters")
			}
			for i := 0; i < len(nums); i++ {
				x := nums[i]
				if isRelative {
					x += current.X
				}
				current = shapes.Point{X: x, Y: current.Y}
				pt := current
				currentSubpath = append(currentSubpath, &pt)
			}
			lastCubicControl = nil
			lastQuadControl = nil
		case 'V':
			if len(nums) < 1 {
				return errors.New("invalid vertical lineto parameters")
			}
			for i := 0; i < len(nums); i++ {
				y := nums[i]
				if isRelative {
					y += current.Y
				}
				current = shapes.Point{X: current.X, Y: y}
				pt := current
				currentSubpath = append(currentSubpath, &pt)
			}
			lastCubicControl = nil
			lastQuadControl = nil
		case 'C':
			if len(nums) < 6 || len(nums)%6 != 0 {
				return errors.New("invalid cubic bezier parameters")
			}
			for i := 0; i < len(nums); i += 6 {
				c1 := shapes.Point{X: nums[i], Y: nums[i+1]}
				c2 := shapes.Point{X: nums[i+2], Y: nums[i+3]}
				end := shapes.Point{X: nums[i+4], Y: nums[i+5]}
				if isRelative {
					c1.X += current.X
					c1.Y += current.Y
					c2.X += current.X
					c2.Y += current.Y
					end.X += current.X
					end.Y += current.Y
				}
				bezier := shapes.CubicBezier{Start: current, Control1: c1, Control2: c2, End: end}
				currentSubpath = append(currentSubpath, &bezier)
				current = end
				lastCubicControl = &c2
				lastQuadControl = nil
			}
		case 'S':
			if len(nums) < 4 || len(nums)%4 != 0 {
				return errors.New("invalid smooth cubic bezier parameters")
			}
			for i := 0; i < len(nums); i += 4 {
				var c1 shapes.Point
				if (lastCmdUpper == 'C' || lastCmdUpper == 'S') && lastCubicControl != nil {
					c1 = shapes.Point{X: 2*current.X - lastCubicControl.X, Y: 2*current.Y - lastCubicControl.Y}
				} else {
					c1 = current
				}
				c2 := shapes.Point{X: nums[i], Y: nums[i+1]}
				end := shapes.Point{X: nums[i+2], Y: nums[i+3]}
				if isRelative {
					c2.X += current.X
					c2.Y += current.Y
					end.X += current.X
					end.Y += current.Y
				}
				bezier := shapes.CubicBezier{Start: current, Control1: c1, Control2: c2, End: end}
				currentSubpath = append(currentSubpath, &bezier)
				current = end
				lastCubicControl = &c2
				lastQuadControl = nil
			}
		case 'Q':
			if len(nums) < 4 || len(nums)%4 != 0 {
				return errors.New("invalid quadratic bezier parameters")
			}
			for i := 0; i < len(nums); i += 4 {
				ctrl := shapes.Point{X: nums[i], Y: nums[i+1]}
				end := shapes.Point{X: nums[i+2], Y: nums[i+3]}
				if isRelative {
					ctrl.X += current.X
					ctrl.Y += current.Y
					end.X += current.X
					end.Y += current.Y
				}
				bezier := shapes.QuadraticBezier{Start: current, Control: ctrl, End: end}
				currentSubpath = append(currentSubpath, &bezier)
				current = end
				lastQuadControl = &ctrl
				lastCubicControl = nil
			}
		case 'T':
			if len(nums) < 2 || len(nums)%2 != 0 {
				return errors.New("invalid smooth quadratic bezier parameters")
			}
			for i := 0; i < len(nums); i += 2 {
				var ctrl shapes.Point
				if (lastCmdUpper == 'Q' || lastCmdUpper == 'T') && lastQuadControl != nil {
					ctrl = shapes.Point{X: 2*current.X - lastQuadControl.X, Y: 2*current.Y - lastQuadControl.Y}
				} else {
					ctrl = current
				}
				end := shapes.Point{X: nums[i], Y: nums[i+1]}
				if isRelative {
					end.X += current.X
					end.Y += current.Y
				}
				bezier := shapes.QuadraticBezier{Start: current, Control: ctrl, End: end}
				currentSubpath = append(currentSubpath, &bezier)
				current = end
				lastQuadControl = &ctrl
				lastCubicControl = nil
			}
		case 'A':
			if len(nums) < 7 || len(nums)%7 != 0 {
				return errors.New("invalid arc parameters")
			}
			for i := 0; i < len(nums); i += 7 {
				rx := nums[i]
				ry := nums[i+1]
				phi := nums[i+2]
				largeArcFlag := nums[i+3] != 0
				sweepFlag := nums[i+4] != 0
				end := shapes.Point{X: nums[i+5], Y: nums[i+6]}
				if isRelative {
					end.X += current.X
					end.Y += current.Y
				}
				if rx == 0 || ry == 0 {
					current = end
					currentSubpath = append(currentSubpath, &end)
					lastCubicControl = nil
					lastQuadControl = nil
					continue
				}
				cubics := arcToCubicBeziers(current, end, rx, ry, phi, largeArcFlag, sweepFlag)
				for j := range cubics {
					currentSubpath = append(currentSubpath, &cubics[j])
				}
				current = end
				lastCubicControl = nil
				lastQuadControl = nil
			}
		case 'Z':
			if current.X != subpathStart.X || current.Y != subpathStart.Y {
				current = subpathStart
				pt := current
				currentSubpath = append(currentSubpath, &pt)
			}
			lastCubicControl = nil
			lastQuadControl = nil
		default:
			return errors.New("unsupported path command")
		}

		lastCmdUpper = upper
		return nil
	}

	for i := 0; i < len(dString); {
		c := dString[i]
		if isPathCommand(c) {
			if err := flush(); err != nil {
				return shapes.Path{}, err
			}
			cmd = c
			nums = nums[:0]
			i++
			continue
		}

		if isNumberStart(c) {
			value, next, ok := parsePathNumber(dString, i)
			if !ok {
				return shapes.Path{}, errors.New("invalid number in path data")
			}
			if cmd == 0 {
				return shapes.Path{}, errors.New("path data missing command")
			}
			nums = append(nums, value)
			i = next
			continue
		}
		i++
	}

	if err := flush(); err != nil {
		return shapes.Path{}, err
	}

	if len(currentSubpath) > 0 {
		subpaths = append(subpaths, currentSubpath)
	}

	return shapes.NewCompoundPath(subpaths), nil
}

func isPathCommand(b byte) bool {
	switch b {
	case 'M', 'm', 'L', 'l', 'H', 'h', 'V', 'v', 'C', 'c', 'S', 's', 'Q', 'q', 'T', 't', 'A', 'a', 'Z', 'z':
		return true
	default:
		return false
	}
}

func isNumberStart(b byte) bool {
	return (b >= '0' && b <= '9') || b == '-' || b == '+' || b == '.'
}

func parsePathNumber(s string, start int) (float64, int, bool) {
	i := start
	if i >= len(s) {
		return 0, start, false
	}
	if s[i] == '+' || s[i] == '-' {
		i++
	}
	hasDigits := false
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		hasDigits = true
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			hasDigits = true
		}
	}
	if !hasDigits {
		return 0, start, false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		expDigits := false
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
			expDigits = true
		}
		if expDigits {
			i = j
		}
	}
	value, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil {
		return 0, start, false
	}
	return value, i, true
}

func arcToCubicBeziers(start, end shapes.Point, rx, ry, xAxisRotation float64, largeArc, sweep bool) []shapes.CubicBezier {
	if start.X == end.X && start.Y == end.Y {
		return nil
	}

	rx = math.Abs(rx)
	ry = math.Abs(ry)
	if rx == 0 || ry == 0 {
		return nil
	}

	phi := xAxisRotation * (math.Pi / 180.0)
	sinPhi, cosPhi := math.Sincos(phi)

	dx := (start.X - end.X) / 2
	dy := (start.Y - end.Y) / 2

	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	rx2 := rx * rx
	ry2 := ry * ry
	x1p2 := x1p * x1p
	y1p2 := y1p * y1p

	lambda := x1p2/rx2 + y1p2/ry2
	if lambda > 1 {
		scale := math.Sqrt(lambda)
		rx *= scale
		ry *= scale
		rx2 = rx * rx
		ry2 = ry * ry
	}

	sign := 1.0
	if largeArc == sweep {
		sign = -1.0
	}

	numerator := rx2*ry2 - rx2*y1p2 - ry2*x1p2
	denominator := rx2*y1p2 + ry2*x1p2
	coef := 0.0
	if denominator != 0 {
		coef = sign * math.Sqrt(math.Max(0, numerator/denominator))
	}

	cxp := coef * (rx * y1p / ry)
	cyp := coef * (-ry * x1p / rx)

	cx := cosPhi*cxp - sinPhi*cyp + (start.X+end.X)/2
	cy := sinPhi*cxp + cosPhi*cyp + (start.Y+end.Y)/2

	vx1 := (x1p - cxp) / rx
	vy1 := (y1p - cyp) / ry
	vx2 := (-x1p - cxp) / rx
	vy2 := (-y1p - cyp) / ry

	theta1 := math.Atan2(vy1, vx1)
	delta := math.Atan2(vx1*vy2-vy1*vx2, vx1*vx2+vy1*vy2)

	if !sweep && delta > 0 {
		delta -= 2 * math.Pi
	} else if sweep && delta < 0 {
		delta += 2 * math.Pi
	}

	segments := int(math.Ceil(math.Abs(delta) / (math.Pi / 2)))
	if segments < 1 {
		segments = 1
	}
	deltaSeg := delta / float64(segments)

	cubics := make([]shapes.CubicBezier, 0, segments)
	for i := 0; i < segments; i++ {
		t1 := theta1 + float64(i)*deltaSeg
		t2 := t1 + deltaSeg

		sin1, cos1 := math.Sincos(t1)
		sin2, cos2 := math.Sincos(t2)
		alpha := (4.0 / 3.0) * math.Tan((t2-t1)/4.0)

		p1 := mapArcPoint(cos1, sin1, rx, ry, cosPhi, sinPhi, cx, cy)
		p2 := mapArcPoint(cos2, sin2, rx, ry, cosPhi, sinPhi, cx, cy)
		c1 := mapArcPoint(cos1-alpha*sin1, sin1+alpha*cos1, rx, ry, cosPhi, sinPhi, cx, cy)
		c2 := mapArcPoint(cos2+alpha*sin2, sin2-alpha*cos2, rx, ry, cosPhi, sinPhi, cx, cy)

		cubics = append(cubics, shapes.CubicBezier{Start: p1, Control1: c1, Control2: c2, End: p2})
	}

	return cubics
}

func mapArcPoint(x, y, rx, ry, cosPhi, sinPhi, cx, cy float64) shapes.Point {
	xr := x * rx
	yr := y * ry
	xRot := cosPhi*xr - sinPhi*yr
	yRot := sinPhi*xr + cosPhi*yr
	return shapes.Point{X: cx + xRot, Y: cy + yRot}
}
