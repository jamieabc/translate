package translator

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	gcp "cloud.google.com/go/translate"

	"golang.org/x/text/language"
)

const (
	gcpAPIRateLimit = 10   // max rate per second
	gcpPayloadLimit = 2000 // max total characters per request
	gcpWordLimit    = 128  // max words per request
	targetLang      = "zh-TW"
	outputFileName  = "out.txt"
)

type googleCloud struct {
	sourceFileName string
}

func (g *googleCloud) Translate() error {
	ctx := context.Background()

	// googleCloud client
	client, err := gcp.NewClient(ctx)
	if nil != err {
		return fmt.Errorf("failed to create googleCloud client: %s", err)
	}

	// setup translation target language
	lang, err := language.Parse(targetLang)
	if nil != err {
		return fmt.Errorf("fail to set Translate destination language: %s", err)
	}

	outputFile, err := os.Create(outputFileName)
	if nil != err {
		return fmt.Errorf("fail to create output file %s with error: %s", outputFileName, err)
	}
	defer outputFile.Close()

	data, err := ioutil.ReadFile(g.sourceFileName)
	if nil != err {
		return fmt.Errorf("read file %s with error: %s", sourceFileName, err)
	}

	err = query(data, outputFile, client, lang, ctx)

	return nil
}

func query(data []byte, output *os.File, client *gcp.Client, lang language.Tag, ctx context.Context) error {
	sleepTimeMillisecond := float64(1000) / gcpAPIRateLimit

	startIndex := 0
	for startIndex < len(data)-1 {
		bs, nextIndex := truncateWords(data, startIndex)
		strs := strings.Split(string(bs), "\n")
		startIndex = nextIndex

		// fmt.Printf("strings to Translate: %s\n\n", strs)

		// Translate
		translated, err := client.Translate(ctx, strs, lang, nil)
		if nil != err {
			return fmt.Errorf("api query error: %v", err)
		}

		time.Sleep(time.Duration(sleepTimeMillisecond) * time.Millisecond)

		for _, s := range translated {
			_, err = output.WriteString(s.Text)
			if nil != err {
				return fmt.Errorf("write file with error: %s", err)
			}
		}
		err = output.Sync()
		if nil != err {
			return fmt.Errorf("sync file %s with error: %s", outputFileName, err)
		}
	}

	return nil
}

func truncateWords(bs []byte, startIndex int) ([]byte, int) {
	var endIndex, segment, selectedLength int
	length := len(bs[startIndex:])

	if length <= gcpPayloadLimit {
		endIndex = startIndex + length - 1
		selectedLength = length
	} else {
		endIndex = startIndex + gcpPayloadLimit - 1
		selectedLength = gcpPayloadLimit - 1
	}

	txt := bs[startIndex:endIndex]
	var i, j int
	for i = 0; i < selectedLength && segment < gcpWordLimit; i++ {
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

	if segment == gcpWordLimit {
		endIndex = startIndex + i + 1
	}

	offsetIdx := bytes.LastIndex(bs[startIndex:endIndex+1], []byte{'\n'})
	return bs[startIndex : startIndex+offsetIdx+1], startIndex + offsetIdx + 1
}

func newGCP(sourceFileName string) Translator {
	return &googleCloud{
		sourceFileName: sourceFileName,
	}
}
