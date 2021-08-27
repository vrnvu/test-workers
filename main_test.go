package main

import "testing"

func TestPipeline(t *testing.T) {
	expected := []result{
		{"1.txt", "a"},
		{"10.txt", "a"},
		{"11.txt", "a"},
		{"13.txt", "a"},
		{"14.txt", "a"},
		{"15.txt", "a"},
		{"16.txt", "a"},
		{"17.txt", "a"},
		{"18.txt", "a"},
		{"19.txt", "a"},
		{"2.txt", "a"},
		{"20.txt", "a"},
		{"3.txt", "a"},
		{"5.txt", "a"},
		{"6.txt", "a"},
	}
	numberExecutions := 10
	for i := 0; i < numberExecutions; i++ {
		actual := GetResult()

		if len(actual) != len(expected) {
			t.Fatalf("invalid len, expected %v, actual %v", expected, actual)
		}
		for i := range expected {
			a := actual[i]
			e := expected[i]
			if a.fileName != e.fileName {
				t.Fatalf("expected %v, actual %v", e, a)
			}
		}

	}
}
