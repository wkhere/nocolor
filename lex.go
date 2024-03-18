package nocolor

// lexer API

type tokenType int

const (
	tokenError tokenType = -1
	tokenAny             = 0
	tokenColor           = 1
)

type token struct {
	typ tokenType
	val []byte
	err error
}

func lexTokens(input []byte, bufSize int) <-chan token {
	l := &lexer{
		input:  input,
		tokens: make(chan token, bufSize),
	}
	go l.run()
	return l.tokens
}

// engine

type lexer struct {
	input      []byte
	start, pos int
	lastw      int
	tokens     chan token
}

type stateFn func(*lexer) stateFn

type lexError string

func (e lexError) Error() string { return string(e) }

func (l *lexer) run() {
	for st := lexStart; st != nil; {
		st = st(l)
	}
	close(l.tokens)
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{typ: t, val: l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) emitError(text string) {
	l.tokens <- token{
		typ: tokenError, val: l.input[l.start:l.pos],
		err: lexError(text),
	}
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
		l.emit(tokenColor)
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
			l.emit(tokenAny)
		}
		l.unbackup()
		l.emitError("binary data")
		return nil
	}
	l.emit(tokenAny)
	return lexStart
}

// predicates; note we consider only ascii runes

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
