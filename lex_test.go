package nocolor

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"testing"
)

type ts []token
type token []byte

func (t token) String() string { return fmt.Sprint(string(t)) }

func t(s string) token { return token([]byte(s)) }

type testWriter struct {
	res ts
}

func (tw *testWriter) Write(p []byte) (int, error) {
	tw.res = append(tw.res, token(p))
	return len(p), nil
}

func eq1(t1, t2 token) bool {
	return bytes.Equal(t1, t2)
}

func eq(ts1, ts2 ts) bool {
	return slices.EqualFunc(ts1, ts2, eq1)
}

var tabLex = []struct {
	data    string
	want    ts
	wantErr error
}{
	{"", ts{}, nil},
	{"aaa", ts{t("aaa")}, nil},
	{"1234", ts{t("1234")}, nil},
	{"#123", ts{t("#123")}, nil},
	{"123#", ts{t("123#")}, nil},

	{"\033[34;40m1234\033[0m",
		ts{t("1234")}, nil},
	{"\033[48;5;17m\033[38;5;19m1234\033[0m",
		ts{t("1234")}, nil},
	{"aaa\033[0mbbb",
		ts{t("aaa"), t("bbb")}, nil},
	{"\0330000",
		ts{t("\0330000")}, nil},
	{"\033[00x0000",
		ts{t("\033[00x0000")}, nil},

	{"\x00",
		ts{t("\x00")}, binErr},
	{"\x00rest",
		ts{t("\x00")}, binErr},
	{" \x00",
		ts{t(" "), t("\x00")}, binErr},
	{"aaa\x00",
		ts{t("aaa"), t("\x00")}, binErr},
	{"111\x00222",
		ts{t("111"), t("\x00")}, binErr},
}

var binErr = fmt.Errorf("binary data")

func TestLex(t *testing.T) {
	for i, tc := range tabLex {
		tw := testWriter{}

		err := lexAndWrite(&tw, []byte(tc.data))

		switch wantErr := tc.wantErr; {
		case wantErr != nil && err == nil:
			t.Errorf("tc[%d] want error %q but have none", i, wantErr)
		case wantErr == nil && err != nil:
			t.Errorf("tc[%d] have unexpected error: %v", i, err)
		case wantErr != nil && err != nil:
			if wantErr.Error() != err.Error() {
				t.Errorf("tc[%d] error mismatch:\nhave: %v\nwant: %v",
					i, err, wantErr,
				)
			}
		case wantErr == nil && err == nil:
			if !eq(tw.res, tc.want) {
				t.Errorf("tc[%d] mismatch\nhave %v\nwant %v", i, tw.res, tc.want)
			}
		}
	}
}

func BenchmarkLex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range tabLex[:10] {
			lexAndWrite(io.Discard, []byte(tc.data))
		}
	}
}

func FuzzNoEmptyTokens(f *testing.F) {
	for _, tc := range tabLex {
		f.Add(tc.data)
	}
	f.Fuzz(func(t *testing.T, s string) {
		tw := testWriter{}

		lexAndWrite(&tw, []byte(s))

		for _, tok := range tw.res {
			if len(tok) == 0 {
				t.Error("empty token:", tok)
			}
		}
	})
}
