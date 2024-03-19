package nocolor

import "io"

// lexer API

func lexAndWrite(w io.Writer, input []byte) error {
	l := &lexer{
		input: input,
		w:     w,
	}
	l.run()
	return l.err
}

// engine

type lexer struct {
	input      []byte
	start, pos int
	lastw      int

	w   io.Writer
	err error
}

type stateFn func(*lexer) stateFn

type lexError string

func (e lexError) Error() string { return string(e) }

func (l *lexer) run() {
	for st := lexStart; st != nil; {
		st = st(l)
	}
}

func (l *lexer) noemit() {
	l.start = l.pos
}

func (l *lexer) emit() {
	// no error check - assuming it is *bufio.Writer:
	l.w.Write(l.input[l.start:l.pos])
	l.start = l.pos
}

func (l *lexer) emitError(text string) {
	l.err = lexError(text)
	l.start = l.pos
}

// input-consuming primitives

const (
	cEOF rune = -1
	cBin      = 0
	cEsc      = '\033'
)

func (l *lexer) readc() (c rune) {
	if len(l.input[l.pos:]) == 0 {
		l.lastw = 0
		return cEOF
	}
	// Note: not using utf8.DecodeRune seems almost like cheating,
	// but in fact non-ascii runes are never used in the state functions
	// below, so it's ok.
	// Still, the rune (not byte) is returned to allow extra values, like -1.
	c = rune(l.input[l.pos])
	l.lastw = 1
	l.pos++
	return c
}

// backup can be used only once after each readc.
func (l *lexer) backup() {
	l.pos -= l.lastw
}

func (l *lexer) unbackup() {
	l.pos += l.lastw
}

// func (l *lexer) peek() rune {
// 	c := l.readc()
// 	l.backup()
// 	return c
// }

// input-consuming helpers

func (l *lexer) acceptOne(c rune) bool {
	if l.readc() == c {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(pred func(rune) bool) {
	for pred(l.readc()) {
	}
	l.backup()
}

func (l *lexer) skipUntil(pred func(rune) bool) {
	for {
		if c := l.readc(); c == cEOF || pred(c) {
			break
		}
	}
	l.backup()
}

// state functions

func lexStart(l *lexer) stateFn {
	switch c := l.readc(); {
	case c == cEOF:
		return nil
	case c == cEsc:
		return lexColorSeq
	default:
		l.backup()
		return lexAny
	}
}

func lexColorSeq(l *lexer) stateFn {
	if l.acceptOne('[') {
		return lexColorValues
	}
	return lexAny
}

func lexColorValues(l *lexer) stateFn {
	l.acceptRun(isDigit)
	switch l.readc() {
	case ';':
		return lexColorValues
	case 'm':
		l.noemit()
		return lexStart
	default:
		return lexAny
	}
}

func lexAny(l *lexer) stateFn {
	var bin bool
	l.skipUntil(func(c rune) bool {
		switch {
		case c == cBin:
			bin = true
			return true
		case c == cEsc:
			return true
		default:
			return false
		}
	})
	if bin {
		if l.pos > l.start {
			l.emit()
		}
		l.unbackup()
		l.emitError("binary data")
		return nil
	}
	l.emit()
	return lexStart
}

// predicates; note we consider only ascii runes

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
