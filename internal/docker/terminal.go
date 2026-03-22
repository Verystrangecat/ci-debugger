package docker

import "golang.org/x/term"

func makeRaw(fd int) (*term.State, error) {
	return term.MakeRaw(fd)
}

func restoreTerminal(fd int, state *term.State) {
	_ = term.Restore(fd, state)
}
