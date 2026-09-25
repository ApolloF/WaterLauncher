package pad

import (
	"testing"
	"time"
)

// SDL loads from the embedded, hash-checked DLL and initialises.
func TestSDLLoads(t *testing.T) {
	s, err := loadSDL()
	if err != nil {
		t.Fatal(err)
	}
	if r, _, _ := s.init.Call(initGamepad); !ok(r) {
		t.Fatalf("SDL_Init: %s", s.errorText())
	}
	s.quit.Call()
}

func TestStartStop(t *testing.T) {
	states := make(chan State, 8)
	m := Start(func(string, bool) {}, func(st State) { states <- st })
	time.Sleep(300 * time.Millisecond)
	m.Stop()
	if st := m.State(); st.Error != "" {
		t.Errorf("state error: %s", st.Error)
	}
}

func TestSetLightValidates(t *testing.T) {
	m := &Manager{cmds: make(chan func(*sdl), 1)}
	for _, bad := range []string{"", "red", "#12345", "#gggggg", "123456#"} {
		if m.SetLight(bad) == nil {
			t.Errorf("SetLight(%q) accepted", bad)
		}
	}
	if err := m.SetLight("#3ea8eb"); err != nil {
		t.Error(err)
	}
}
