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

type result struct {
	fileName string
	value    string
}

func main() {
	r := GetResult(15)
	fmt.Println(r)
}

func GetResult(limit int) []result {

	g, ctx := errgroup.WithContext(context.Background())

	names := getFileNames()

	inC := make(chan result)
	outC := make(chan result)
	doneC := make(chan bool, 1)

	// coordinator
	g.Go(func() error {
		from := 0
		to := 2 // not safe should verify len(names) xd
		max := len(names)
		for {
			select {
			case <-doneC:
				close(inC)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				// generate a group of workers
				// what about this? we starting a new group each batch iteration
				sg, sctx := errgroup.WithContext(ctx)
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if from == max {
						close(inC)
						return nil
					}

					for i := from; i < to; i++ {
						n := names[i]
						processor := processValue(n, inC, sctx)
						sg.Go(processor)
					}

					from, to = nextStep(from, to, max)

					if err := sg.Wait(); err != nil {
						panic(err)
					}

				}

			}
		}
	})

	g.Go(func() error {
		// blocking inC
		count := 0
		toSignal := true
		for r := range inC {
			// we send anyway but signal our coordinator that we have already enough good data
			if toSignal && count >= limit {
				doneC <- true
				toSignal = false
			}
			// attempt to send result to outC, verify ctx
			select {
			case <-ctx.Done():
				return ctx.Err()
			// problem this is a blocking operation, so we need a buffer
			case outC <- r:
			}

			// only count after success sent
			count += 1
			fmt.Println(count)
		}
		// TODO what happens if inC is closed at this point?
		return nil
	})

	// we want to block main, we are blocking our coordinator
	// so we are waiting for the doneC to be completed
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

func buildResults(resultC <-chan result, limit int) []result {
	var results []result
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

func nextStep(from, to, max int) (int, int) {
	step := 2
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

func processValue(fileName string, out chan<- result, ctx context.Context) func() error {
	return func() error {
		v := getValue(fileName)
		if v != "a" {
			return nil
		}
		select {
		case out <- result{fileName: fileName, value: v}:
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
