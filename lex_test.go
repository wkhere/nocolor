package nocolor

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

type b = []byte
type ts = []token

func t(typ tokenType, val []byte) token { return token{typ: typ, val: val} }
func te(err error, val []byte) token {
	return token{typ: tokenError, val: val, err: err}
}

func eq1(t1, t2 token) bool {
	if t1.typ != t2.typ {
		return false
	}
	if !bytes.Equal(t1.val, t2.val) {
		return false
	}
	if t1.err == nil && t2.err == nil {
		return true
	}
	return fmt.Sprint(t1.err) == fmt.Sprint(t2.err)
}

func eq(tt1, tt2 []token) bool {
	return slices.EqualFunc(tt1, tt2, eq1)
}

func (t token) String() string {
	var s string
	switch t.typ {
	case tokenError:
		s = "tokenError"
	case tokenAny:
		s = "tokenAny"
	case tokenColor:
		s = "tokenColor"
	default:
		s = "!!wrong token type!!"
	}
	if t.err == nil {
		return fmt.Sprintf("{%d:%s %q}", t.typ, s, t.val)
	}
	return fmt.Sprintf("{%d:%s %q, %q}", t.typ, s, t.val, t.err)
}

var tabLex = []struct {
	data string
	want []token
}{
	{"", ts{}},
	{"aaa", ts{t(tokenAny, b("aaa"))}},
	{"1234", ts{t(tokenAny, b("1234"))}},
	{"#123", ts{t(tokenAny, b("#123"))}},
	{"123#", ts{t(tokenAny, b("123#"))}},

	{"\033[34;40m1234\033[0m", ts{
		t(tokenColor, b("\033[34;40m")),
		t(tokenAny, b("1234")),
		t(tokenColor, b("\033[0m")),
	}},
	{"\033[48;5;17m\033[38;5;19m1234\033[0m", ts{
		t(tokenColor, b("\033[48;5;17m")),
		t(tokenColor, b("\033[38;5;19m")),
		t(tokenAny, b("1234")),
		t(tokenColor, b("\033[0m")),
	}},
	{"aaa\033[0mbbb", ts{
		t(tokenAny, b("aaa")),
		t(tokenColor, b("\033[0m")),
		t(tokenAny, b("bbb")),
	}},
	{"\0330000", ts{
		t(tokenAny, b("\0330000")),
	}},
	{"\033[00x0000", ts{
		t(tokenAny, b("\033[00x0000")),
	}},

	{"\x00", ts{te(binErr, b("\x00"))}},
	{"\x00rest", ts{te(binErr, b("\x00"))}},
	{" \x00", ts{t(tokenAny, b(" ")), te(binErr, b("\x00"))}},
	{"aaa\x00", ts{t(tokenAny, b("aaa")), te(binErr, b("\x00"))}},
	{"111\x00222", ts{t(tokenAny, b("111")), te(binErr, b("\x00"))}},
}

var binErr = fmt.Errorf("binary data")

func TestLex(t *testing.T) {
	for i, tc := range tabLex {
		have := collect(lexTokens(b(tc.data), 4))
		if !eq(have, tc.want) {
			t.Errorf("tc[%d] mismatch\nhave %v\nwant %v", i, have, tc.want)
		}
	}
}

func BenchmarkLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range tabLex[:10] {
			collect(lexTokens([]byte(tc.data), 4))
		}
	}
}

func collect[T any](ch <-chan T) (a []T) {
	for t := range ch {
		a = append(a, t)
	}
	return
}
