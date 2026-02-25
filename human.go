// human.go
package human

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Cursor wraps a rod.Page and tracks the virtual cursor position.
type Cursor struct {
	page  *rod.Page
	pos   point
	viewW int
	viewH int
	cfg   Config
}

// New creates a Cursor for the given page with optional configuration.
func New(page *rod.Page, opts ...Option) *Cursor {
	cfg := DefaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	w, h := getViewportSize(page)
	return &Cursor{
		page:  page,
		pos:   point{0, 0},
		viewW: w,
		viewH: h,
		cfg:   cfg,
	}
}

// Move moves the cursor to a random point inside the element.
func (c *Cursor) Move(el *rod.Element) error {
	dst, err := scrollIntoView(el)
	if err != nil {
		return err
	}
	c.pos = moveMouse(c.page, c.pos, dst, c.viewW, c.viewH, &c.cfg)
	return nil
}

// MoveSteady moves to the element with a straight trajectory.
func (c *Cursor) MoveSteady(el *rod.Element) error {
	dst, err := scrollIntoView(el)
	if err != nil {
		return err
	}
	cfg := c.cfg
	cfg.Steadiness = 1.0
	c.pos = moveMouse(c.page, c.pos, dst, c.viewW, c.viewH, &cfg)
	return nil
}

// MoveToPoint moves the cursor to exact coordinates.
func (c *Cursor) MoveToPoint(x, y float64) {
	dst := point{x, y}
	c.pos = moveMouse(c.page, c.pos, dst, c.viewW, c.viewH, &c.cfg)
}

// Click moves to the element and left-clicks.
func (c *Cursor) Click(el *rod.Element) error {
	if err := c.Move(el); err != nil {
		return err
	}
	c.click(1, 0)
	return nil
}

// DoubleClick moves to the element and double-clicks.
func (c *Cursor) DoubleClick(el *rod.Element) error {
	if err := c.Move(el); err != nil {
		return err
	}
	c.click(2, 0)
	return nil
}

// ClickHold clicks and holds for the given duration in milliseconds.
func (c *Cursor) ClickHold(el *rod.Element, holdMs int) error {
	if err := c.Move(el); err != nil {
		return err
	}
	c.click(1, holdMs)
	return nil
}

// RightClick moves to the element and right-clicks.
func (c *Cursor) RightClick(el *rod.Element) error {
	if err := c.Move(el); err != nil {
		return err
	}
	mouseDown(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonRight)
	sleepJitter(50, 120)
	mouseUp(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonRight)
	sleepJitter(c.cfg.ClickDwell[0], c.cfg.ClickDwell[1])
	return nil
}

// DragDrop drags from one element to another.
func (c *Cursor) DragDrop(from, to *rod.Element) error {
	if err := c.Move(from); err != nil {
		return err
	}
	mouseDown(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonLeft)
	sleepJitter(80, 200)

	dst, err := scrollIntoView(to)
	if err != nil {
		mouseUp(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonLeft)
		return err
	}
	c.pos = moveMouse(c.page, c.pos, dst, c.viewW, c.viewH, &c.cfg)
	sleepJitter(50, 120)
	mouseUp(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonLeft)
	sleepJitter(c.cfg.ClickDwell[0], c.cfg.ClickDwell[1])
	return nil
}

// Scroll dispatches a mouse wheel event. Positive deltaY scrolls down.
func (c *Cursor) Scroll(deltaX, deltaY float64) {
	_ = proto.InputDispatchMouseEvent{
		Type:   proto.InputDispatchMouseEventTypeMouseWheel,
		X:      c.pos.X,
		Y:      c.pos.Y,
		DeltaX: deltaX,
		DeltaY: deltaY,
	}.Call(c.page)
	sleepJitter(100, 300)
}

func (c *Cursor) click(n int, holdMs int) {
	for i := 0; i < n; i++ {
		mouseDown(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonLeft)
		if holdMs > 0 {
			time.Sleep(time.Duration(holdMs) * time.Millisecond)
		} else {
			sleepJitter(c.cfg.ClickHold[0], c.cfg.ClickHold[1])
		}
		mouseUp(c.page, c.pos.X, c.pos.Y, proto.InputMouseButtonLeft)
		sleepJitter(c.cfg.ClickDwell[0], c.cfg.ClickDwell[1])
	}
}

// Type types text with human-like per-key delays.
func (c *Cursor) Type(text string) {
	typeText(c.page, text, &c.cfg)
}

// TypeWithSpeed types text with custom per-key delay range in milliseconds.
func (c *Cursor) TypeWithSpeed(text string, minDelayMs, maxDelayMs int) {
	cfg := c.cfg
	cfg.TypeDelay = [2]int{minDelayMs, maxDelayMs}
	typeText(c.page, text, &cfg)
}

// PressKey presses and releases a single key.
func (c *Cursor) PressKey(key input.Key) {
	pressKey(c.page, key)
}

// KeyCombo presses a key combination (e.g. Ctrl+A).
func (c *Cursor) KeyCombo(modifiers []input.Key, key input.Key) {
	pressKeyCombo(c.page, modifiers, key)
}

// Pos returns the current cursor position.
func (c *Cursor) Pos() (float64, float64) {
	return c.pos.X, c.pos.Y
}

// ClickAndType clicks on an element, clears any existing value, then types the
// given text with human-like keystroke timing.
func (c *Cursor) ClickAndType(el *rod.Element, text string) error {
	if err := c.Click(el); err != nil {
		return err
	}

	c.KeyCombo([]input.Key{input.ControlLeft}, input.KeyA)
	sleepJitter(40, 120)
	c.PressKey(input.Backspace)
	sleepJitter(60, 180)

	// Verify the field is actually empty via DOM property, if not, clear with JS and re-fire events to ensure any framework bindings update
	val, err := el.Property("value")
	if err == nil && val.String() != "" {
		_, _ = el.Eval(`() => {
			this.value = '';
			this.dispatchEvent(new Event('input',  {bubbles: true}));
			this.dispatchEvent(new Event('change', {bubbles: true}));
		}`)
		sleepJitter(30, 80)
	}

	c.Type(text)

	sleepJitter(300, 1000)
	return nil
}
