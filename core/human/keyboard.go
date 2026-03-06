// keyboard.go
package human

import (
	"math/rand/v2"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

func typeText(page *rod.Page, text string, cfg *Config) {
	minDelay := cfg.TypeDelay[0]
	maxDelay := cfg.TypeDelay[1]
	if minDelay <= 0 {
		minDelay = 50
	}
	if maxDelay <= 0 {
		maxDelay = 190
	}

	prevWasSpace := false
	for _, ch := range text {
		typeChar(page, ch)
		ms := minDelay + rand.IntN(maxDelay-minDelay+1)

		if prevWasSpace {
			ms += rand.IntN(80) + 30
		}
		if ch == '.' || ch == ',' || ch == '!' || ch == '?' {
			ms += rand.IntN(150) + 60
		}

		if rand.Float64() < cfg.ThinkPause {
			ms += rand.IntN(400) + 150
		}

		prevWasSpace = ch == ' '
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func typeChar(page *rod.Page, ch rune) {
	if unicode.IsPrint(ch) {
		_ = proto.InputDispatchKeyEvent{
			Type: proto.InputDispatchKeyEventTypeChar,
			Text: string(ch),
		}.Call(page)
		return
	}

	switch ch {
	case '\n', '\r':
		pressKey(page, input.Enter)
	case '\t':
		pressKey(page, input.Tab)
	}
}

func pressKey(page *rod.Page, key input.Key) {
	down := key.Encode(proto.InputDispatchKeyEventTypeKeyDown, 0)
	up := key.Encode(proto.InputDispatchKeyEventTypeKeyUp, 0)
	_ = down.Call(page)
	_ = up.Call(page)
}

func pressKeyCombo(page *rod.Page, modifiers []input.Key, key input.Key) {
	modFlag := 0
	for _, mod := range modifiers {
		modFlag |= mod.Modifier()
		down := mod.Encode(proto.InputDispatchKeyEventTypeKeyDown, modFlag)
		_ = down.Call(page)
	}

	down := key.Encode(proto.InputDispatchKeyEventTypeKeyDown, modFlag)
	up := key.Encode(proto.InputDispatchKeyEventTypeKeyUp, modFlag)
	_ = down.Call(page)
	_ = up.Call(page)

	for i := len(modifiers) - 1; i >= 0; i-- {
		modFlag &^= modifiers[i].Modifier()
		upMod := modifiers[i].Encode(proto.InputDispatchKeyEventTypeKeyUp, modFlag)
		_ = upMod.Call(page)
	}
}
