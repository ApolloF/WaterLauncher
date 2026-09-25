// Package pad reads game controllers through SDL3 and turns them into the
// few actions the big picture interface understands (move, confirm, back,
// menu, …), with key repeat for held directions. It also drives rumble and
// the DualSense lightbar. SDL runs on a thread of its own and is only ever
// touched from there.
package pad

import (
	"encoding/binary"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// Actions sent to the interface.
const (
	Up      = "up"
	Down    = "down"
	Left    = "left"
	Right   = "right"
	Confirm = "confirm" // ✕ / A
	Back    = "back"    // ○ / B
	Action  = "action"  // □ / X
	Info    = "info"    // △ / Y
	Menu    = "menu"    // Options / Menu
	View    = "view"    // Create / View
	Home    = "home"    // PS / Xbox button
	LB      = "lb"
	RB      = "rb"
	LT      = "lt"
	RT      = "rt"
)

// Kind is a controller family, for the button glyphs.
type Kind string

const (
	PlayStation Kind = "playstation"
	Xbox        Kind = "xbox"
	Nintendo    Kind = "nintendo"
	Other       Kind = "other"
)

// State describes the controller in use.
type State struct {
	Connected bool   `json:"connected"`
	Name      string `json:"name"`
	Kind      Kind   `json:"kind"`
	DualSense bool   `json:"dualSense"`
	Battery   int    `json:"battery"` // percent, -1 unknown
	Wireless  bool   `json:"wireless"`
	Error     string `json:"error,omitempty"`
}

// SDL3 constants used here.
const (
	initGamepad = 0x00002000

	evQuit           = 0x100
	evGamepadAxis    = 0x650
	evGamepadDown    = 0x651
	evGamepadUp      = 0x652
	evGamepadAdded   = 0x653
	evGamepadRemoved = 0x654

	typePS3 = 4
	typePS4 = 5
	typePS5 = 6
	typeSw1 = 7

	connWireless = 2
)

// SDL gamepad buttons → actions.
var buttons = map[uint8]string{
	0: Confirm, 1: Back, 2: Action, 3: Info, 4: View, 5: Home, 6: Menu,
	9: LB, 10: RB, 11: Up, 12: Down, 13: Left, 14: Right,
}

const (
	repeatDelay = 380 * time.Millisecond
	repeatEvery = 115 * time.Millisecond
	stickOn     = 16000 // of 32767: a stick counts as pressed past this
	stickOff    = 11000 // and released below this
	triggerOn   = 16000
)

// Manager owns the SDL thread.
type Manager struct {
	onAction func(action string, repeat bool)
	onState  func(State)

	cmds chan func(*sdl)
	quit chan struct{}
	done chan struct{}

	mu    sync.Mutex
	state State
}

// Start loads SDL and begins reading controllers. onAction gets every
// action (repeat=true for auto-repeat of a held direction); onState gets
// controller changes. Both are called from the SDL thread.
func Start(onAction func(string, bool), onState func(State)) *Manager {
	m := &Manager{onAction: onAction, onState: onState, cmds: make(chan func(*sdl), 16),
		quit: make(chan struct{}), done: make(chan struct{})}
	m.state.Battery = -1
	go m.loop()
	return m
}

// State returns the current controller state.
func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Stop shuts SDL down.
func (m *Manager) Stop() {
	select {
	case <-m.quit:
	default:
		close(m.quit)
	}
	<-m.done
}

// Rumble plays a short effect: "tick" (moving), "confirm" or "error".
func (m *Manager) Rumble(effect string) {
	var lo, hi uint16
	var ms uint32
	switch effect {
	case "tick":
		lo, hi, ms = 0, 0x2800, 22
	case "confirm":
		lo, hi, ms = 0x3800, 0x5000, 55
	case "error":
		lo, hi, ms = 0x7000, 0x2000, 140
	default:
		return
	}
	m.do(func(s *sdl) {
		if gp := m.current(); gp != 0 {
			s.rumbleGamepad.Call(gp, uintptr(lo), uintptr(hi), uintptr(ms))
		}
	})
}

// SetLight sets the DualSense lightbar to a #rrggbb colour.
func (m *Manager) SetLight(hex string) error {
	if len(hex) != 7 || hex[0] != '#' {
		return errors.New("colour must be #rrggbb")
	}
	v, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return err
	}
	r, g, b := uintptr(v>>16&0xff), uintptr(v>>8&0xff), uintptr(v&0xff)
	m.do(func(s *sdl) {
		if gp := m.current(); gp != 0 {
			s.setGamepadLED.Call(gp, r, g, b)
		}
	})
	return nil
}

func (m *Manager) do(fn func(*sdl)) {
	select {
	case m.cmds <- fn:
	default: // the SDL thread is busy; a dropped effect doesn't matter
	}
}

// ---- the SDL thread ----

type gamepad struct {
	ptr  uintptr
	name string
	kind Kind
	ds   bool
}

var (
	padsMu  sync.Mutex
	pads    = map[uint32]*gamepad{}
	current uint32 // the controller last used
)

func (m *Manager) current() uintptr {
	padsMu.Lock()
	defer padsMu.Unlock()
	if gp := pads[current]; gp != nil {
		return gp.ptr
	}
	return 0
}

func (m *Manager) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(m.done)

	s, err := loadSDL()
	if err != nil {
		m.setState(func(st *State) { st.Error = "controller support unavailable: " + err.Error() })
		<-m.quit
		return
	}
	// WaterLauncher has no SDL window, so SDL must deliver input no matter
	// which window has focus.
	for k, v := range map[string]string{
		"SDL_JOYSTICK_ALLOW_BACKGROUND_EVENTS": "1",
		"SDL_JOYSTICK_HIDAPI_PS5_PLAYER_LED":   "0",
	} {
		s.setHint.Call(uintptr(unsafe.Pointer(cstr(k))), uintptr(unsafe.Pointer(cstr(v))))
	}
	if r, _, _ := s.init.Call(initGamepad); !ok(r) {
		m.setState(func(st *State) { st.Error = "controller support unavailable: " + s.errorText() })
		<-m.quit
		return
	}
	defer s.quit.Call()

	ev := make([]byte, 128)
	held := map[string]time.Time{} // direction → next repeat time
	axes := map[uint8]int16{}
	battery := time.NewTicker(30 * time.Second)
	defer battery.Stop()
	tick := time.NewTicker(8 * time.Millisecond)
	defer tick.Stop()

	press := func(a string) {
		if a == "" {
			return
		}
		m.onAction(a, false)
		if a == Up || a == Down || a == Left || a == Right {
			held[a] = time.Now().Add(repeatDelay)
		}
	}
	release := func(a string) { delete(held, a) }

	for {
		select {
		case <-m.quit:
			m.closeAll(s)
			return
		case fn := <-m.cmds:
			fn(s)
			continue
		case <-battery.C:
			m.refreshState(s)
			continue
		case <-tick.C:
		}
		for {
			r, _, _ := s.pollEvent.Call(uintptr(unsafe.Pointer(&ev[0])))
			if !ok(r) {
				break
			}
			typ := binary.LittleEndian.Uint32(ev[0:])
			which := binary.LittleEndian.Uint32(ev[16:])
			switch typ {
			case evGamepadAdded:
				m.open(s, which)
			case evGamepadRemoved:
				m.close(s, which)
			case evGamepadDown:
				m.use(s, which)
				press(buttons[ev[20]])
			case evGamepadUp:
				release(buttons[ev[20]])
			case evGamepadAxis:
				m.use(s, which)
				axis := ev[20]
				v := int16(binary.LittleEndian.Uint16(ev[24:]))
				prev := axes[axis]
				axes[axis] = v
				switch axis {
				case 0, 1: // left stick, with hysteresis so a wobbly stick doesn't stutter
					neg, pos := Left, Right
					if axis == 1 {
						neg, pos = Up, Down
					}
					stick := func(a string, beyond, inside bool) {
						if _, isHeld := held[a]; beyond && !isHeld {
							press(a)
						} else if inside && isHeld {
							release(a)
						}
					}
					stick(neg, v < -stickOn, v > -stickOff)
					stick(pos, v > stickOn, v < stickOff)
				case 4, 5: // triggers
					a := LT
					if axis == 5 {
						a = RT
					}
					if v > triggerOn && prev <= triggerOn {
						m.onAction(a, false)
					}
				}
			case evQuit:
			}
		}
		now := time.Now()
		for a, next := range held {
			if now.After(next) {
				m.onAction(a, true)
				held[a] = now.Add(repeatEvery)
			}
		}
	}
}

func (m *Manager) open(s *sdl, id uint32) {
	r, _, _ := s.openGamepad.Call(uintptr(id))
	if r == 0 {
		return
	}
	name, _, _ := s.gamepadName.Call(r)
	t, _, _ := s.gamepadType.Call(r)
	gp := &gamepad{ptr: r, name: gostr(name)}
	switch int32(t) {
	case typePS3, typePS4, typePS5:
		gp.kind = PlayStation
	case typeSw1, 8, 9, 10:
		gp.kind = Nintendo
	case 2, 3:
		gp.kind = Xbox
	default:
		if strings.Contains(strings.ToLower(gp.name), "xbox") {
			gp.kind = Xbox
		} else {
			gp.kind = Other
		}
	}
	gp.ds = int32(t) == typePS5
	padsMu.Lock()
	pads[id] = gp
	current = id
	padsMu.Unlock()
	m.refreshState(s)
}

func (m *Manager) close(s *sdl, id uint32) {
	padsMu.Lock()
	gp := pads[id]
	delete(pads, id)
	if current == id {
		current = 0
		for other := range pads {
			current = other
			break
		}
	}
	padsMu.Unlock()
	if gp != nil {
		s.closeGamepad.Call(gp.ptr)
	}
	m.refreshState(s)
}

func (m *Manager) closeAll(s *sdl) {
	padsMu.Lock()
	defer padsMu.Unlock()
	for id, gp := range pads {
		s.closeGamepad.Call(gp.ptr)
		delete(pads, id)
	}
	current = 0
}

// use makes the controller that was just touched the current one.
func (m *Manager) use(s *sdl, id uint32) {
	padsMu.Lock()
	changed := current != id && pads[id] != nil
	if changed {
		current = id
	}
	padsMu.Unlock()
	if changed {
		m.refreshState(s)
	}
}

func (m *Manager) refreshState(s *sdl) {
	padsMu.Lock()
	gp := pads[current]
	padsMu.Unlock()
	m.setState(func(st *State) {
		*st = State{Battery: -1}
		if gp == nil {
			return
		}
		st.Connected, st.Name, st.Kind, st.DualSense = true, gp.name, gp.kind, gp.ds
		var pct int32 = -1
		s.gamepadPowerInfo.Call(gp.ptr, uintptr(unsafe.Pointer(&pct)))
		st.Battery = int(pct)
		c, _, _ := s.gamepadConnectionState.Call(gp.ptr)
		st.Wireless = int32(c) == connWireless
	})
}

func (m *Manager) setState(fn func(*State)) {
	m.mu.Lock()
	fn(&m.state)
	st := m.state
	m.mu.Unlock()
	if m.onState != nil {
		m.onState(st)
	}
}
