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

		err = lexAndWrite(bw, line)
		if err != nil {
			return err
		}
		if err = bw.Flush(); err != nil {
			return err
		}
	}
	return nil
}
