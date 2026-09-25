package platform

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"unsafe"

	"golang.org/x/sys/windows"
)

var reSecretName = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

func secretPath(name string) (string, error) {
	if !reSecretName.MatchString(name) {
		return "", errors.New("bad secret name")
	}
	return filepath.Join(AppDir(), "secrets", name+".bin"), nil
}

// SaveSecret stores a secret (an API key) encrypted with Windows DPAPI,
// so only this Windows user on this PC can read it back. An empty value
// deletes it.
func SaveSecret(name, value string) error {
	p, err := secretPath(name)
	if err != nil {
		return err
	}
	if value == "" {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	in := []byte(value)
	var out windows.DataBlob
	if err := windows.CryptProtectData(&windows.DataBlob{Size: uint32(len(in)), Data: &in[0]}, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	enc := append([]byte(nil), unsafe.Slice(out.Data, out.Size)...)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, enc, 0o600)
}

// LoadSecret reads a secret stored with SaveSecret; "" when there is none.
func LoadSecret(name string) string {
	p, err := secretPath(name)
	if err != nil {
		return ""
	}
	enc, err := os.ReadFile(p)
	if err != nil || len(enc) == 0 {
		return ""
	}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&windows.DataBlob{Size: uint32(len(enc)), Data: &enc[0]}, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return ""
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return string(unsafe.Slice(out.Data, out.Size))
}
