package search

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
)

//Result результат поиска где искомая фраза, строка, номер строки и позиция начало фразы в строке
type Result struct {
	Phrase  string
	Line    string
	LineNum int64
	ColNum  int64
}

//All ищет все вхождение phrase в текстовых файлах files
func All(ctx context.Context, phrase string, files []string) <-chan []Result {
	ch := make(chan []Result)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)

	for i := 0; i < len(files); i++ {
		wg.Add(1)

		go func(ctx context.Context, path string, ch chan<- []Result) {
			defer wg.Done()

			results := []Result{}
			data, err := os.ReadFile(path)
			if err != nil {
				log.Print("Can`t open the file:", err)
			}

			dataStr := string(data)
			splitData := strings.Split(dataStr, "\n")

			for index, line := range splitData {
				if strings.Contains(line, phrase) {
					result := Result{
						Phrase:  phrase,
						Line:    line,
						LineNum: int64(index + 1),
						ColNum:  int64(strings.Index(line, phrase) + 1),
					}
					results = append(results, result)
				}
			}

			if len(results) > 0 {
				ch <- results
			}
		}(ctx, files[i], ch)
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	cancel()
	return ch

}

func Any(ctx context.Context, phrase string, files []string) <-chan Result {
	ch := make(chan Result)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)

	for i := 0; i < len(files); i++ {
		wg.Add(1)

		go func(ctx context.Context, path string, ch chan<- Result) {
			defer wg.Done()

			data, err := os.ReadFile(path)
			if err != nil {
				log.Print("Can`t open the file:", err)
			}

			dataStr := string(data)
			splitData := strings.Split(dataStr, "\n")

			for index, line := range splitData {
				select {
				case <-ctx.Done():
					return
				default:
					if strings.Contains(line, phrase) {
						result := Result{
							Phrase:  phrase,
							Line:    line,
							LineNum: int64(index + 1),
							ColNum:  int64(strings.Index(line, phrase) + 1),
						}
						ch <- result
						cancel()
					}
				}
			}
		}(ctx, files[i], ch)
	}

	go func() {
		defer close(ch)
		wg.Wait()
		cancel()
	}()

	return ch
}
