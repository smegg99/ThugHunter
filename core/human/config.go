// config.go
package human

// Config controls cursor timing and behavior.
type Config struct {
	// Direct bypasses all human-like mouse movement, click timing, and
	// keystroke delays. When true, interactions use rod's native element
	// methods directly. It will also use JS-based versions of some methods to avoid issues in headless mode.
	Direct bool

	// Hesitation is the probability of pausing before mouse movement (0-1).
	Hesitation float64

	// MicroPause is the probability of a brief pause mid-movement (0-1).
	MicroPause float64

	// Steadiness controls curve straightness (0 = wobbly, 1 = straight).
	Steadiness float64

	// ClickHold is the [min, max] ms to hold a mouse button during click.
	ClickHold [2]int

	// ClickDwell is the [min, max] ms to pause after a click.
	ClickDwell [2]int

	// TypeDelay is the [min, max] ms between keystrokes.
	TypeDelay [2]int

	// ThinkPause is the probability of a longer pause while typing (0-1).
	ThinkPause float64
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Hesitation: 0.6,
		MicroPause: 0.3,
		Steadiness: 0,
		ClickHold:  [2]int{50, 130},
		ClickDwell: [2]int{120, 350},
		TypeDelay:  [2]int{50, 190},
		ThinkPause: 0.08,
	}
}

// Option configures a Cursor.
type Option func(*Config)

// WithConfig replaces the entire configuration.
func WithConfig(cfg Config) Option {
	return func(c *Config) { *c = cfg }
}

// WithHesitation sets the pre-move pause probability (0-1).
func WithHesitation(chance float64) Option {
	return func(c *Config) { c.Hesitation = chance }
}

// WithMicroPause sets the mid-move pause probability (0-1).
func WithMicroPause(chance float64) Option {
	return func(c *Config) { c.MicroPause = chance }
}

// WithSteadiness sets curve straightness (0 = wobbly, 1 = straight).
func WithSteadiness(s float64) Option {
	return func(c *Config) { c.Steadiness = s }
}

// WithClickTiming sets click hold and post-click dwell ranges in ms.
func WithClickTiming(holdMin, holdMax, dwellMin, dwellMax int) Option {
	return func(c *Config) {
		c.ClickHold = [2]int{holdMin, holdMax}
		c.ClickDwell = [2]int{dwellMin, dwellMax}
	}
}

// WithTypingSpeed sets the per-key delay range in ms.
func WithTypingSpeed(minMs, maxMs int) Option {
	return func(c *Config) { c.TypeDelay = [2]int{minMs, maxMs} }
}

// WithThinkPause sets the probability of longer pauses while typing (0-1).
func WithThinkPause(chance float64) Option {
	return func(c *Config) { c.ThinkPause = chance }
}

// Fast returns a configuration with fatser movements and typing.
func Fast() Option {
	return func(c *Config) {
		c.Hesitation = 0.15
		c.MicroPause = 0.08
		c.Steadiness = 0.6
		c.ClickHold = [2]int{25, 60}
		c.ClickDwell = [2]int{50, 120}
		c.TypeDelay = [2]int{15, 55}
		c.ThinkPause = 0.02
	}
}

// Direct returns a configuration that bypasses all human-like simulation
// and uses rod's native element methods. Guaranteed to work in headless.
func Direct() Option {
	return func(c *Config) {
		c.Direct = true
	}
}

// WithDirect sets the direct mode flag.
func WithDirect(on bool) Option {
	return func(c *Config) {
		c.Direct = on
	}
}

// Swift returns a configuration that is faster than average but still human-like.
func Swift() Option {
	return func(c *Config) {
		c.Hesitation = 0.3
		c.MicroPause = 0.15
		c.Steadiness = 0.4
		c.ClickHold = [2]int{35, 80}
		c.ClickDwell = [2]int{80, 200}
		c.TypeDelay = [2]int{30, 90}
		c.ThinkPause = 0.04
	}
}

// Beginner returns a configuration with more human-like imperfections.
func Beginner() Option {
	return func(c *Config) {
		c.Hesitation = 0.9
		c.MicroPause = 0.6
		c.Steadiness = 0.0
		c.ClickHold = [2]int{100, 250}
		c.ClickDwell = [2]int{300, 700}
		c.TypeDelay = [2]int{150, 500}
		c.ThinkPause = 0.25
	}
}

// Casual returns a configuration that mimics a casual user.
func Casual() Option {
	return func(c *Config) {
		c.Hesitation = 0.5
		c.MicroPause = 0.25
		c.Steadiness = 0.2
		c.ClickHold = [2]int{50, 130}
		c.ClickDwell = [2]int{120, 350}
		c.TypeDelay = [2]int{60, 200}
		c.ThinkPause = 0.08
	}
}
