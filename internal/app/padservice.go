package app

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/ApolloF/WaterLauncher/internal/launch"
	"github.com/ApolloF/WaterLauncher/internal/logx"
	"github.com/ApolloF/WaterLauncher/internal/pad"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Controller events for the interface.
const (
	EventPadAction = "pad:action"
	EventPadState  = "pad:state"
)

// PadAction is one controller action.
type PadAction struct {
	Action string `json:"action"`
	Repeat bool   `json:"repeat"`
}

func init() {
	application.RegisterEvent[PadAction](EventPadAction)
	application.RegisterEvent[pad.State](EventPadState)
}

// PadService is the controller: actions, glyph family, rumble and lightbar.
type PadService struct {
	c   *Core
	mu  sync.Mutex
	mgr *pad.Manager
	// The interface asks for rumble on every move; cap it so a held
	// direction doesn't turn into one long buzz.
	lastRumble time.Time
}

// NewPadService binds the controller layer to core.
func NewPadService(c *Core) *PadService { return &PadService{c: c} }

// ServiceStartup starts reading controllers.
func (s *PadService) ServiceStartup(context.Context, application.ServiceOptions) error {
	s.mgr = pad.Start(s.onAction, s.onState)
	s.c.pad.Store(s.mgr)
	return nil
}

// ServiceShutdown releases the controllers.
func (s *PadService) ServiceShutdown() error {
	if s.mgr != nil {
		s.mgr.Stop()
	}
	return nil
}

func (s *PadService) onAction(action string, repeat bool) {
	home := action == pad.Home && !repeat && s.c.Settings.Get().PSButton
	// While a game runs the controller belongs to the game: actions only
	// drive the overlay, which the PS / Xbox button opens and closes.
	if s.c.Launch.Current().Phase == launch.Running {
		if home {
			s.c.shell.ToggleOverlay()
		} else if s.c.shell.OverlayOpen() {
			s.c.emit(EventOverlayAction, PadAction{Action: action, Repeat: repeat})
		}
		return
	}
	// Otherwise the PS / Xbox button brings WaterLauncher forward, from
	// wherever the user is.
	if home {
		s.c.shell.OpenMain()
	}
	s.c.emit(EventPadAction, PadAction{Action: action, Repeat: repeat})
}

func (s *PadService) onState(st pad.State) {
	if st.Error != "" {
		logx.Printf("controller: %s", st.Error)
	}
	s.c.emit(EventPadState, st)
}

// State returns the controller in use.
func (s *PadService) State() pad.State {
	if s.mgr == nil {
		return pad.State{Battery: -1}
	}
	return s.mgr.State()
}

// Rumble plays a short effect ("tick", "confirm", "error") when the user
// has haptics on.
func (s *PadService) Rumble(effect string) {
	if s.mgr == nil || !s.c.Settings.Get().Haptics {
		return
	}
	s.mu.Lock()
	if effect == "tick" && time.Since(s.lastRumble) < 60*time.Millisecond {
		s.mu.Unlock()
		return
	}
	s.lastRumble = time.Now()
	s.mu.Unlock()
	s.mgr.Rumble(effect)
}

var reHex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// SetLight tints the DualSense lightbar (#rrggbb) when the user has that on.
func (s *PadService) SetLight(hex string) {
	if s.mgr == nil || !s.c.Settings.Get().Lightbar || !reHex.MatchString(hex) {
		return
	}
	_ = s.mgr.SetLight(hex)
}
