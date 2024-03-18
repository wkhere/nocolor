package nocolor

import (
	"bufio"
	"io"
)

func Run(r io.Reader, w io.Writer) error {
	br, bw := bufio.NewReaderSize(r, 4096), bufio.NewWriter(w)

	for {
		line, err := br.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF && len(line) == 0 {
			break
		}

		err = procLine(bw, line)
		if err != nil {
			return err
		}
		if err = bw.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func procLine(bw *bufio.Writer, input []byte) error {
	for token := range lexTokens(input, estTokensNum(input)) {
		switch token.typ {
		case tokenError:
			return token.err
		case tokenColor:
			continue
		default:
			bw.Write(token.val)
		}
	}
	return nil
}

func estTokensNum(input []byte) int {
	switch n := len(input); {
	case n > 320:
		return 16
	default:
		return 4 * n / 80
	}
}
