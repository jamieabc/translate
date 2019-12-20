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
	ctx            context.Context
	client         *gcp.Client
	lang           language.Tag
	output         *os.File
	content        []byte
}

func (g *googleCloud) Initialise() error {
	var err error
	g.ctx = context.Background()

	// googleCloud client
	g.client, err = gcp.NewClient(g.ctx)
	if nil != err {
		return fmt.Errorf("failed to create googleCloud client: %s", err)
	}

	// setup translation target language
	g.lang, err = language.Parse(targetLang)
	if nil != err {
		return fmt.Errorf("fail to set Translate destination language: %s", err)
	}

	g.output, err = os.Create(outputFileName)
	if nil != err {
		return fmt.Errorf("fail to create output file %s with error: %s", outputFileName, err)
	}

	g.content, err = ioutil.ReadFile(g.sourceFileName)
	if nil != err {
		return fmt.Errorf("read file %s with error: %s", g.sourceFileName, err)
	}

	return nil
}

func (g *googleCloud) Translate() error {
	return g.query()
}

func (g *googleCloud) query() error {
	sleepTimeMillisecond := float64(1000) / gcpAPIRateLimit

	startIndex := 0
	for startIndex < len(g.content)-1 {
		bs, nextIndex := truncateWords(g.content, startIndex)
		strs := strings.Split(string(bs), "\n")
		startIndex = nextIndex

		// fmt.Printf("strings to Translate: %s\n\n", strs)

		// Translate
		translated, err := g.client.Translate(g.ctx, strs, g.lang, nil)
		if nil != err {
			return fmt.Errorf("api query error: %v", err)
		}

		time.Sleep(time.Duration(sleepTimeMillisecond) * time.Millisecond)

		for _, s := range translated {
			_, err = g.output.WriteString(s.Text)
			if nil != err {
				return fmt.Errorf("write file with error: %s", err)
			}
		}
		err = g.output.Sync()
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
