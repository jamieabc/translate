package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	fileName, err := checkArgs(args)

	if nil != err {
		fmt.Printf("\n%s\n\n", err.Error())
		return
	}

	content, err := ioutil.ReadFile(fileName)
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
	lan, err := language.Parse(targteLang)
	if nil != err {
		log.Fatalf("Failed to parse target language: %v", err)
	}

	translateStartIdx := 0
	str := bytes.NewBuffer(content).String()
	f, err := os.Create(outputFile)
	if nil != err {
		log.Fatalf("Failed to create output file %s: %v", outputFile, err)
	}
	defer f.Close()

	for {

		// data := strings.Split(str, "\n")
		subStr, translateEndIdx := truncateWords(str, translateStartIdx)
		data := strings.Split(subStr, "\n")

		// translate
		translation, err := client.Translate(ctx, data, lan, nil)
		if nil != err {
			log.Fatalf("Failed to translate text: %v", err)
		}

		for _, str := range translation {
			f.WriteString(str.Text)
			f.WriteString("\n")
		}
		f.Sync()

		if translateEndIdx >= len(str) {
			break
		} else {
			translateStartIdx = translateEndIdx
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
