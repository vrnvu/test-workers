package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Result holds the filename and a value, this value in our example is always a
type Result struct {
	fileName string
	value    string
}

func main() {
	r := GetResult(15)
	fmt.Println(r)
}

// GetResult will return the list of files that contain the "a" value inside sorted by the
// fileName and limited to limit parameter
func GetResult(limit int) []Result {

	g, ctx := errgroup.WithContext(context.Background())

	names := getFileNames()

	inC := make(chan Result)
	outC := make(chan Result)
	doneC := make(chan bool, 1)

	g.Go(func() error {
		step := 5
		from, to, max := 0, step, len(names)
		for {
			select {
			case <-doneC:
				close(inC)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				if from == max {
					close(inC)
					return nil
				}

				sg, sctx := errgroup.WithContext(ctx)

				for i := from; i < to; i++ {
					n := names[i]
					processor := processValue(sctx, n, inC)
					sg.Go(processor)
				}

				if err := sg.Wait(); err != nil {
					return err
				}

				from, to = nextStep(step, from, to, max)
			}

		}
	})

	g.Go(func() error {
		count := 0
		toSignal := true
		for r := range inC {
			// we send anyway but signal our coordinator that we have already enough good data
			// we use a toSignal to not block and proceed to send the rest of inC to have them all
			if toSignal && count >= limit {
				doneC <- true
				toSignal = false
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case outC <- r:
			}

			count++
			fmt.Println(count)
		}
		return nil
	})

	// we do not want to block main, we are waiting on our coordinator group
	// if success we close our outC been used in buildResults
	go func() {
		if err := g.Wait(); err != nil {
			panic(err)
		}
		fmt.Println("finished")
		close(outC)
	}()

	r := buildResults(outC, limit)
	return r
}

func buildResults(resultC <-chan Result, limit int) []Result {
	var results []Result
	for r := range resultC {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].fileName < results[j].fileName
	})

	if limit > len(results) {
		limit = len(results)
	}
	return results[:limit]
}

func nextStep(step, from, to, max int) (int, int) {
	nfrom := from + step

	if nfrom > max {
		nfrom = max
	}

	nto := to + step

	if nto > max {
		nto = max
	}

	return nfrom, nto
}

func processValue(ctx context.Context, fileName string, out chan<- Result) func() error {
	return func() error {
		v := getValue(fileName)
		if v != "a" {
			return nil
		}
		select {
		case out <- Result{fileName: fileName, value: v}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}
}

func getValue(fileName string) string {
	b, err := os.ReadFile("test/" + fileName)
	if err != nil {
		panic(err)
	}
	value := string(b)
	value = strings.TrimSpace(value)
	return value
}

func getFileNames() []string {
	files, err := ioutil.ReadDir("test")
	if err != nil {
		panic(err)
	}
	var results []string
	for _, f := range files {
		results = append(results, f.Name())
	}
	sort.Strings(results)
	return results
}
