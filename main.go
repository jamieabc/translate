package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

const (
	apiRequestLimit  = 2000 // max total characters per request
	textSegmentLimit = 128  // max words per request
	apiRatePerSecond = 10   // aip requests / second
	targetLang       = "zh-TW"
	outputFile       = "out.txt"
)

func printHelp() {
	fmt.Printf("\nUsage: translate file_name\n")
}

func main() {
	args := os.Args[1:]
	sourceFile, err := checkArgs(args)

	if nil != err {
		fmt.Printf("\n%s\n\n", err.Error())
		return
	}

	ctx := context.Background()

	// create a client
	client, err := translate.NewClient(ctx)
	if nil != err {
		log.Fatalf("Failed to create client: %v", err)
	}

	// get text to translate
	lan, err := language.Parse(targetLang)
	if nil != err {
		log.Fatalf("Failed to parse target language: %v", err)
	}

	destFile, err := os.Create(outputFile)
	if nil != err {
		log.Fatalf("Failed to create output file %s: %v", outputFile, err)
	}
	defer destFile.Close()

	data, err := ioutil.ReadFile(sourceFile)
	if nil != err {
		fmt.Printf("read file with error: %s\n", err)
		return
	}

	sleepTimeMS := float64(1000) / apiRatePerSecond

	startIndex := 0
	for startIndex < len(data)-1 {
		bs, nextIndex := truncateWords(data, startIndex)
		strs := strings.Split(string(bs), "\n")
		startIndex = nextIndex

		// fmt.Printf("strings to translate: %s\n\n", strs)

		// translate
		translated, err := client.Translate(ctx, strs, lan, nil)
		if nil != err {
			log.Fatalf("error: %v", err)
		}

		time.Sleep(time.Duration(sleepTimeMS) * time.Millisecond)

		for _, s := range translated {
			destFile.WriteString(s.Text)
			destFile.WriteString("\n")
		}
		destFile.Sync()
	}
}

func checkArgs(args []string) (string, error) {
	if 0 == len(args) {
		printHelp()
		return "", fmt.Errorf("error: Insufficient argument")
	}

	fileName := args[0]
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("Error: file '%s' not exist", fileName)
	}
	return fileName, nil
}

func truncateWords(bs []byte, startIndex int) ([]byte, int) {
	var endIndex, segment, selectedLength int
	length := len(bs[startIndex:])

	if length <= apiRequestLimit {
		endIndex = startIndex + length - 1
		selectedLength = length
	} else {
		endIndex = startIndex + apiRequestLimit - 1
		selectedLength = apiRequestLimit - 1
	}

	txt := bs[startIndex:endIndex]
	var i, j int
	for i = 0; i < selectedLength && segment < textSegmentLimit; i++ {
		if txt[i] == ' ' {
			segment++
			for j = i + 1; j < selectedLength; j++ {
				if txt[j] != ' ' {
					break
				}
			}
			i = j
		}
	}

	if segment == textSegmentLimit {
		endIndex = startIndex + i + 1
	}

	offsetIdx := bytes.LastIndex(bs[startIndex:endIndex+1], []byte{'\n'})
	return bs[startIndex : startIndex+offsetIdx+1], startIndex + offsetIdx + 1
}
