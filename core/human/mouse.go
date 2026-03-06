// mouse.go
package human

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func moveMouse(page *rod.Page, origin, dst point, viewW, viewH int, cfg *Config) point {
	opts := randomCurveOpts(viewW, viewH, origin, dst, cfg.Steadiness)
	curve := generateCurve(origin, dst, opts)
	n := len(curve)
	if n == 0 {
		return origin
	}

	dist := math.Hypot(dst.X-origin.X, dst.Y-origin.Y)
	baseDur := 350.0 + math.Log2(dist+1)*80.0
	jitter := baseDur * (0.8 + rand.Float64()*0.4)
	totalMs := math.Max(jitter, 250)

	microPauseIdx := -1
	microPauseMs := 0
	if rand.Float64() < cfg.MicroPause && n > 10 {
		microPauseIdx = n/3 + rand.IntN(n/3)
		microPauseMs = randIntRange(40, 120)
	}

	weights := make([]float64, n)
	var totalWeight float64
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		speed := 0.4 + 0.6*math.Sin(t*math.Pi)
		w := 1.0 / speed
		weights[i] = w
		totalWeight += w
	}

	if rand.Float64() < cfg.Hesitation {
		sleepJitter(50, 180)
	}

	for i, p := range curve {
		_ = proto.InputDispatchMouseEvent{
			Type: proto.InputDispatchMouseEventTypeMouseMoved,
			X:    p.X,
			Y:    p.Y,
		}.Call(page)

		stepMs := (weights[i] / totalWeight) * totalMs
		stepMs *= 0.85 + rand.Float64()*0.30
		if stepMs < 0.5 {
			stepMs = 0.5
		}
		time.Sleep(time.Duration(stepMs*1000) * time.Microsecond)

		if i == microPauseIdx {
			time.Sleep(time.Duration(microPauseMs) * time.Millisecond)
		}
	}

	sleepJitter(30, 90)
	return curve[n-1]
}

func mouseDown(page *rod.Page, x, y float64, button proto.InputMouseButton) {
	_ = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMousePressed,
		X:          x,
		Y:          y,
		Button:     button,
		ClickCount: 1,
	}.Call(page)
}

func mouseUp(page *rod.Page, x, y float64, button proto.InputMouseButton) {
	_ = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMouseReleased,
		X:          x,
		Y:          y,
		Button:     button,
		ClickCount: 1,
	}.Call(page)
}

func scrollIntoView(el *rod.Element) (point, error) {
	// Wait until the element is visible, scrolled into view, and not covered
	// by another element. This is essential for headless mode where layout
	// updates can lag behind DOM mutations.
	pt, err := el.WaitInteractable()
	if err != nil {
		return point{}, err
	}

	sleepJitter(50, 150)

	shape, err := el.Shape()
	if err != nil || shape == nil {
		// Fallback: use the interactable point returned by rod.
		if pt != nil {
			return point{X: pt.X, Y: pt.Y}, nil
		}
		return point{}, fmt.Errorf("element has no shape and no interactable point")
	}
	box := shape.Box()

	// Validate the element has non-zero dimensions.
	if box.Width < 1 || box.Height < 1 {
		if pt != nil {
			return point{X: pt.X, Y: pt.Y}, nil
		}
		return point{}, fmt.Errorf("element has zero-size bounding box: %.0fx%.0f", box.Width, box.Height)
	}

	rx := float64(randIntRange(20, 80)) / 100.0
	ry := float64(randIntRange(20, 80)) / 100.0
	return point{
		X: box.X + box.Width*rx,
		Y: box.Y + box.Height*ry,
	}, nil
}

func getViewportSize(page *rod.Page) (int, int) {
	metrics, err := proto.PageGetLayoutMetrics{}.Call(page)
	if err != nil {
		return 1920, 1080
	}
	w := int(metrics.CSSLayoutViewport.ClientWidth)
	h := int(metrics.CSSLayoutViewport.ClientHeight)
	// In headless mode the layout metrics may report 0; fall back to a
	// reasonable default so curve generation doesn't degenerate.
	if w < 1 || h < 1 {
		return 1920, 1080
	}
	return w, h
}

func sleepJitter(loMs, hiMs int) {
	ms := loMs + rand.IntN(hiMs-loMs+1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
