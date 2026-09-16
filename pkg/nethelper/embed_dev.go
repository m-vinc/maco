//go:build !prod

package nethelper

func embeddedBinary() ([]byte, bool) {
	return nil, false
}
