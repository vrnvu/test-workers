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
	resultC := make(chan result)

	g, ctx := errgroup.WithContext(context.Background())

	names := getFileNames()
	for _, n := range names {
		processor := processValue(n, resultC, ctx)
		g.Go(processor)
	}

	go func() {
		if err := g.Wait(); err != nil {
			if err.Error() != "success" {
				panic(err)
			}
		}
	}()

	results := processResult(resultC)

	sort.Slice(results, func(i, j int) bool {
		return results[i].fileName < results[j].fileName
	})
	fmt.Println(results)
}

func GetResult() []result {
	result := make(chan result)

	g, ctx := errgroup.WithContext(context.Background())

	names := getFileNames()

	for _, n := range names {
		processor := processValue(n, result, ctx)
		g.Go(processor)
	}

	go func() {
		if err := g.Wait(); err != nil {
			panic(err)
		}
	}()

	results := processResult(result)
	sort.Slice(results, func(i, j int) bool {
		return results[i].fileName < results[j].fileName
	})

	return results
}

func processResult(resultC <-chan result) []result {
	var results []result
	count := 1
	for {
		select {
		case r, ok := <-resultC:
			if !ok {
				return results
			}
			if count > 15 {
				return results
			}
			count += 1
			results = append(results, r)
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
