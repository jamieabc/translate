package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

const (
	apiRequestLimit = 7000
	targteLang      = "zh-TW"
	outputFile      = "out.txt"
)

func printHelp() {
	fmt.Printf("\nUsage: translate file_name\n")
}

func checkArgs(args []string) (string, error) {
	if 0 == len(args) {
		printHelp()
		return "", fmt.Errorf("Error: Insufficient argument")
	}

	fileName := args[0]
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("Error: file '%s' not exist", fileName)
	}
	return fileName, nil
}

func main() {
	args := os.Args[1:]
	sourceFile, err := checkArgs(args)

	if nil != err {
		fmt.Printf("\n%s\n\n", err.Error())
		return
	}

	f, err := os.Open(sourceFile)
	if nil != err {
		fmt.Printf("open %s with error: %s\n", sourceFile, err)
		return
	}
	reader := bufio.NewReader(f)

	ctx := context.Background()

	// create a client
	client, err := translate.NewClient(ctx)
	if nil != err {
		log.Fatalf("Failed to create client: %v", err)
	}

	// get text to translate
	lan, err := language.Parse(targteLang)
	if nil != err {
		log.Fatalf("Failed to parse target language: %v", err)
	}

	destFile, err := os.Create(outputFile)
	if nil != err {
		log.Fatalf("Failed to create output file %s: %v", outputFile, err)
	}
	defer destFile.Close()

	exit := false

	for {
		data, err := reader.Peek(apiRequestLimit)
		if 0 != len(data) && nil != err {
			fmt.Printf("read file with error: %s", err)
			exit = true
		}
		strs := strings.Split(string(data), "\n")
		fmt.Printf("strings: %v\n", strs)

		// translate
		translated, err := client.Translate(ctx, strs, lan, nil)
		if nil != err {
			log.Fatalf("Failed to translate text: %v", err)
		}

		for _, s := range translated {
			destFile.WriteString(s.Text)
			destFile.WriteString("\n")
		}
		destFile.Sync()

		if exit {
			break
		}
	}
}

func truncateWords(str string, start int) (string, int) {
	length := len(str)
	if length <= apiRequestLimit {
		return str, length - 1
	}

	var endIdx int
	if start+apiRequestLimit >= length {
		endIdx = length
	} else {
		endIdx = start + apiRequestLimit
	}

	offsetIdx := strings.LastIndex(str[start:endIdx], "\n")
	return str[start : start+offsetIdx], start + offsetIdx + 1
}
