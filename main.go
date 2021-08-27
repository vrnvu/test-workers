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

func GetResult(limit int) []result {
	inC := make(chan result)

	g, ctx := errgroup.WithContext(context.Background())

	names := getFileNames()

	g.Go(func() error {
		return producer(ctx, names, inC)
	})

	// TODO example of worker,  println as they come
	// g.Go(func() error {
	// 	return consumer(ctx, inC)
	// })

	go func() {
		if err := g.Wait(); err != nil {
			panic(err)
		}
	}()

	// TODO process the results, we will potentially have more than 15 entries
	// we first sort all of them, then return top 15 elements in results as our solution

	return buildResults(inC, limit)
}

func buildResults(inC <-chan result, limit int) []result {
	var results []result
	for r := range inC {
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

func producer(ctx context.Context, names []string, inC chan<- result) error {
	sg, sctx := errgroup.WithContext(ctx)
	from := 0
	to := 2 // not safe should verify len(names) xd
	max := len(names)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if from == max {
				if err := sg.Wait(); err != nil {
					panic(err)
				}
				close(inC)
				return nil
			} else {
				for i := from; i < to; i++ {
					n := names[i]
					processor := processValue(n, inC, sctx)
					sg.Go(processor)
				}
				from, to = nextStep(from, to, max)
			}
		}
	}
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

func consumer(ctx context.Context, inC <-chan result) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r, ok := <-inC:
			if !ok {
				return nil
			}
			fmt.Println(r)
		}
	}
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
