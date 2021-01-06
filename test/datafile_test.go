package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/algorithmiaio/algorithmia-go"
)

func TestDataFile(t *testing.T) {
	client := algorithmia.NewClient(os.Getenv("ALGORITHMIA_DEFAULT_API_KEY"), "")
	remoteFile := client.File("data://.my/nonexistant/nonreal")
	f, err := remoteFile.File()
	if err == nil {
		t.Fatal("non-existent file retrieved")
		f.Close()
	} else {
		fmt.Println("error received:", err)
	}

	s, err := remoteFile.StringContents()
	if err == nil {
		t.Fatal("non-existent string contents retrieved", s)
	} else {
		fmt.Println("error received:", err)
	}
}
