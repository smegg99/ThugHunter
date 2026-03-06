// curve.go
package human

import (
	"math"
	"math/rand/v2"

	"github.com/fogleman/ease"
)

type point struct {
	X, Y float64
}

func generateCurve(from, to point, opts curveOpts) []point {
	knots := generateInternalKnots(from, to, opts)
	allPoints := make([]point, 0, len(knots)+2)
	allPoints = append(allPoints, from)
	allPoints = append(allPoints, knots...)
	allPoints = append(allPoints, to)

	midCount := int(math.Max(
		math.Abs(from.X-to.X),
		math.Max(math.Abs(from.Y-to.Y), 2),
	))
	curvePoints := bezierPoints(midCount, allPoints)
	curvePoints = distortPoints(curvePoints, opts.distortionMean, opts.distortionStdDev, opts.distortionFreq)
	curvePoints = tweenPoints(curvePoints, opts.tween, opts.targetPoints)
	return curvePoints
}

type curveOpts struct {
	offsetBoundaryX  int
	offsetBoundaryY  int
	knotsCount       int
	distortionMean   float64
	distortionStdDev float64
	distortionFreq   float64
	tween            func(float64) float64
	targetPoints     int
}

func generateInternalKnots(from, to point, opts curveOpts) []point {
	left := math.Min(from.X, to.X) - float64(opts.offsetBoundaryX)
	right := math.Max(from.X, to.X) + float64(opts.offsetBoundaryX)
	down := math.Min(from.Y, to.Y) - float64(opts.offsetBoundaryY)
	up := math.Max(from.Y, to.Y) + float64(opts.offsetBoundaryY)

	knots := make([]point, opts.knotsCount)
	for i := range knots {
		knots[i] = point{
			X: left + rand.Float64()*(right-left),
			Y: down + rand.Float64()*(up-down),
		}
	}
	return knots
}

func distortPoints(pts []point, mean, stdDev, freq float64) []point {
	if len(pts) <= 2 {
		return pts
	}
	result := make([]point, 0, len(pts))
	result = append(result, pts[0])
	for i := 1; i < len(pts)-1; i++ {
		p := pts[i]
		if rand.Float64() < freq {
			delta := rand.NormFloat64()*stdDev + mean
			p.Y += delta
		}
		result = append(result, p)
	}
	result = append(result, pts[len(pts)-1])
	return result
}

func tweenPoints(pts []point, tween func(float64) float64, targetCount int) []point {
	if targetCount < 2 {
		targetCount = 2
	}
	result := make([]point, targetCount)
	for i := 0; i < targetCount; i++ {
		t := float64(i) / float64(targetCount-1)
		idx := int(tween(t) * float64(len(pts)-1))
		if idx >= len(pts) {
			idx = len(pts) - 1
		}
		result[i] = pts[idx]
	}
	return result
}

func binomial(n, k int) float64 {
	return float64(factorial(n)) / float64(factorial(k)*factorial(n-k))
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	r := 1
	for i := 2; i <= n; i++ {
		r *= i
	}
	return r
}

func bernsteinBasis(t float64, i, n int) float64 {
	return binomial(n, i) * math.Pow(t, float64(i)) * math.Pow(1-t, float64(n-i))
}

func bezierPoints(count int, controlPoints []point) []point {
	if count < 2 {
		count = 2
	}
	n := len(controlPoints) - 1
	result := make([]point, count)
	for i := 0; i < count; i++ {
		t := float64(i) / float64(count-1)
		var x, y float64
		for j, cp := range controlPoints {
			b := bernsteinBasis(t, j, n)
			x += cp.X * b
			y += cp.Y * b
		}
		result[i] = point{x, y}
	}
	return result
}

var defaultTweens = []func(float64) float64{
	ease.OutExpo,
	ease.InOutQuint,
	ease.InOutSine,
	ease.InOutQuart,
	ease.InOutExpo,
	ease.InOutCubic,
	ease.InOutCirc,
	ease.Linear,
	ease.OutSine,
	ease.OutQuart,
	ease.OutQuint,
	ease.OutCubic,
	ease.OutCirc,
}

func randomCurveOpts(viewportW, viewportH int, from, to point, steadiness float64) curveOpts {
	minW := float64(viewportW) * 0.15
	maxW := float64(viewportW) * 0.85
	minH := float64(viewportH) * 0.15
	maxH := float64(viewportH) * 0.85

	tween := defaultTweens[rand.IntN(len(defaultTweens))]

	obxRanges := []struct{ lo, hi int }{{20, 45}, {45, 75}, {75, 100}}
	obxWeights := []float64{0.2, 0.65, 0.15}
	offsetBoundaryX := randWeightedRange(obxRanges, obxWeights)

	obyRanges := []struct{ lo, hi int }{{20, 45}, {45, 75}, {75, 100}}
	obyWeights := []float64{0.2, 0.65, 0.15}
	offsetBoundaryY := randWeightedRange(obyRanges, obyWeights)

	knotsWeights := []float64{0.15, 0.36, 0.17, 0.12, 0.08, 0.04, 0.03, 0.02, 0.015, 0.005}
	knotsCount := weightedChoice(knotsWeights) + 1

	distMean := float64(randIntRange(80, 110)) / 100
	distStdDev := float64(randIntRange(85, 110)) / 100
	distFreq := float64(randIntRange(25, 70)) / 100

	tpRanges := []struct{ lo, hi int }{{120, 180}, {180, 250}, {250, 350}}
	tpWeights := []float64{0.30, 0.50, 0.20}
	targetPoints := randWeightedRange(tpRanges, tpWeights)

	if from.X < minW || from.X > maxW || from.Y < minH || from.Y > maxH {
		offsetBoundaryX = 1
		offsetBoundaryY = 1
		knotsCount = 1
	}
	if to.X < minW || to.X > maxW || to.Y < minH || to.Y > maxH {
		offsetBoundaryX = 1
		offsetBoundaryY = 1
		knotsCount = 1
	}

	if steadiness > 0 {
		s := steadiness
		offsetBoundaryX = int(float64(offsetBoundaryX)*(1-s) + 10*s)
		offsetBoundaryY = int(float64(offsetBoundaryY)*(1-s) + 10*s)
		distMean = distMean*(1-s) + 1.2*s
		distStdDev = distStdDev*(1-s) + 1.2*s
		distFreq = distFreq*(1-s) + 1.0*s
	}

	return curveOpts{
		offsetBoundaryX:  offsetBoundaryX,
		offsetBoundaryY:  offsetBoundaryY,
		knotsCount:       knotsCount,
		distortionMean:   distMean,
		distortionStdDev: distStdDev,
		distortionFreq:   distFreq,
		tween:            tween,
		targetPoints:     targetPoints,
	}
}

func randIntRange(lo, hi int) int {
	return lo + rand.IntN(hi-lo)
}

func weightedChoice(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := rand.Float64() * total
	cum := 0.0
	for i, w := range weights {
		cum += w
		if r <= cum {
			return i
		}
	}
	return len(weights) - 1
}

func randWeightedRange(ranges []struct{ lo, hi int }, weights []float64) int {
	idx := weightedChoice(weights)
	r := ranges[idx]
	return randIntRange(r.lo, r.hi)
}
