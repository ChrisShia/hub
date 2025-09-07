package jsonlog

import (
	"fmt"
	"sync"
	"testing"
)

func Test(t *testing.T) {

	out := make([][]byte, 0)

	w := MockWriter{&out}

	logger := New(w, LevelInfo)

	var wg sync.WaitGroup

	t.Run("", func(t *testing.T) {
		n := 10
		wg.Add(n)
		for i := 1; i <= n; i++ {
			go func(i int) {
				defer wg.Done()
				logger.PrintInfo(fmt.Sprintf("Message %d", i), nil)
			}(i)
		}

		wg.Wait()
		for i := 0; i < len(out); i++ {
			fmt.Println(string(out[i]))
		}
	})

}

type MockWriter struct {
	bs *[][]byte
}

func (m MockWriter) Write(line []byte) (n int, err error) {
	//maybe deal with zeo len bs
	*m.bs = append(*m.bs, line)
	return len(line), nil
}
