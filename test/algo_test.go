package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/algorithmiaio/algorithmia-go"
)

func TestAlgo(t *testing.T) {
	c := algorithmia.NewClient(os.Getenv("ALGORITHMIA_API_KEY"), "")
	algo, err := c.Algo("algo://demo/Hello")
	if err != nil {
		t.Fatal(err)
	}
	r, err := algo.Pipe("Author")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Response: ", r)
	resp, ok := r.(*algorithmia.AlgoResponse)
	if !ok {
		t.Fatal("did not receive an AlgoResponse")
	}

	if r, ok := resp.Result.(string); !ok {
		t.Fatal("string answer expected")
	} else {
		if r != "Hello Author" {
			t.Fatal("wrong result")
		}
	}
}
