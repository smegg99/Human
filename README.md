# Human

Human-like mouse and keyboard for [go-rod](https://go-rod.github.io/).

Moves the cursor along randomized Bezier curves and types with natural timing. Works in headless mode (all input goes through CDP). Includes a **Direct mode** that bypasses all simulation and uses JS-based element methods, guaranteed to work in headless. I needed such functionality in my [ThugHunter](https://github.com/smegg99/ThugHunter) project, but it proved useful in other automation tasks as well so I made it into a separate package. Inspired by [riflosnake/HumanCursor](https://github.com/riflosnake/HumanCursor).

## Install

```
go get github.com/smegg99/human
```

## Usage

Look at the examples in the [tests](./tests) for more details on configuration and usage.

```go
cursor := human.New(page)

cursor.Click(element)
cursor.Type("hello world")
```

### Configuration

```go
// Built-in Presets
cursor := human.New(page, human.Fast())      // power user
cursor := human.New(page, human.Swift())     // fast but natural
cursor := human.New(page, human.Casual())    // everyday user
cursor := human.New(page, human.Beginner())  // inexperienced

// Direct mode bypasses all human-like simulation, uses JS element methods.
// Guaranteed to work in headless mode.
cursor := human.New(page, human.Direct())

// Individual options
cursor := human.New(page,
    human.WithTypingSpeed(20, 60),
    human.WithSteadiness(0.5),
    human.WithHesitation(0.3),
)

// Full config
cursor := human.New(page, human.WithConfig(human.Config{
    Hesitation: 0.4,
    MicroPause: 0.2,
    Steadiness: 0.3,
    ClickHold:  [2]int{40, 100},
    ClickDwell: [2]int{100, 300},
    TypeDelay:  [2]int{30, 120},
    ThinkPause: 0.05,
}))
```

### Mouse

```go
cursor.Move(el)
cursor.MoveSteady(el)
cursor.MoveToPoint(500, 300)
cursor.Click(el)
cursor.DoubleClick(el)
cursor.RightClick(el)
cursor.ClickHold(el, 500)
cursor.DragDrop(from, to)
cursor.Scroll(0, 300)
```

### Keyboard

```go
cursor.Type("text")
cursor.TypeWithSpeed("text", 20, 60)
cursor.PressKey(input.Enter)
cursor.KeyCombo([]input.Key{input.ControlLeft}, input.KeyA)
```

### Click and Type

```go
// Clicks the element, clears existing value, types text
cursor.ClickAndType(el, "hello world")
```

## Check out this discord server!

https://discord.gg/edGmPTFufu

## License

MIT
