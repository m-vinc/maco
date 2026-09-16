package engine

import "github.com/m-vinc/maco/pkg/hostnet"

func (e *Engine) Interfaces() ([]hostnet.Port, error) {
	return hostnet.Ports()
}
