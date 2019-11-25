package test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/algorithmiaio/algorithmia-go"
)

var client1 = algorithmia.NewClient(os.Getenv("ALGORITHMIA_API_KEY"), "")

func TestListFilesWithPaging(t *testing.T) {
	const (
		NUM_FILES = 1100
		EXTENSION = ".txt"
	)
	dd := client1.Dir("data://.my/golangLargeDataDirList")
	if exists, err := dd.Exists(); err != nil {
		t.Fatal(err)
	} else if !exists {
		if err := dd.Create(nil); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < NUM_FILES; i++ {
			fname := fmt.Sprint(i)
			err := dd.File(fname + EXTENSION).Put([]byte(fname))
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	seenFiles := make([]bool, NUM_FILES)
	numFiles := 0

	for f := range dd.Files() {
		if f.Err != nil {
			t.Fatal(f.Err)
		}
		numFiles++
		name, err := f.Object.(*algorithmia.DataFile).Name()
		if err != nil {
			t.Fatal(err)
		}
		index, err := strconv.Atoi(name[:len(name)-len(EXTENSION)])
		if err != nil {
			t.Fatal(err)
		}
		seenFiles[index] = true
	}
	allSeen := true
	for _, cur := range seenFiles {
		allSeen = (allSeen && cur)
	}

	if numFiles != NUM_FILES {
		t.Fatal("numFiles != NUM_FILES")
	}

	if !allSeen {
		t.Fatal("!allSeen")
	}
}
