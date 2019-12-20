package main

import (
	"fmt"
	"os"

	"github.com/jamieabc/translate/pkg/translator"
)

func main() {
	if len(os.Args) == 1 {
		printHelp()
		return
	}

	gcp, err := translator.NewTranslator(translator.GCP, os.Args[1])
	if nil != err {
		fmt.Printf("new google cloud translator with error: %s", err)
		return
	}

	err = gcp.Translate()
	if nil != err {
		fmt.Printf("google cloud translate with error: %s", err)
		return
	}

	return
}

func printHelp() {
	fmt.Println("please input filename")
	fmt.Println("usage: translate [file name]")
}
