package translator

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
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

type googleCloudAPI interface {
	Translate(context.Context, []string, language.Tag, *gcp.Options) ([]gcp.Translation, error)
}

type googleCloud struct {
	sourceFileName string
	ctx            context.Context
	client         googleCloudAPI
	lang           language.Tag
	output         *os.File
	content        []byte
}

func (g *googleCloud) Initialise(sourceFileName string) error {
	var err error
	g.sourceFileName = sourceFileName

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
	sleepTimeMillisecond := float64(1000) / gcpAPIRateLimit

	startIndex := 0
	var bs []byte
	for startIndex < len(g.content)-1 {
		bs, startIndex = translatedWords(g.content, startIndex)

		// Translate
		translated, err := g.client.Translate(g.ctx, []string{string(bs)}, g.lang, nil)
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

	fmt.Printf("total translated: %d\n", len(g.content))

	return nil
}

func translatedWords(content []byte, startIndex int) ([]byte, int) {
	var index int
	length := len(content)

	// use 128 words as limitation, since 2000 characters is unlikely to be reached.
	space := 0
	for index = startIndex + 1; index < length && space < gcpWordLimit; index++ {
		if content[index] == ' ' {
			space++
		}
	}

	// check if max characters reached
	if index-startIndex+1 > gcpPayloadLimit {
		index = startIndex + gcpPayloadLimit - 1
	}

	// return if length reached
	if index >= length-1 {
		return content[startIndex:], length
	}

	offsetIndex := bytes.LastIndex(content[startIndex:index+1], []byte{'.'})

	return content[startIndex : startIndex+offsetIndex+1], startIndex + offsetIndex + 1
}

func newGCP(args []interface{}) (Translator, error) {
	ctx := context.Background()

	// setup translation target language
	lang, err := language.Parse(targetLang)
	if nil != err {
		return nil, fmt.Errorf("fail to set Translate destination language: %s", err)
	}

	// googleCloud client
	var client googleCloudAPI
	if len(args) > 0 {
		client = args[0].(googleCloudAPI)
	} else {
		client, err = gcp.NewClient(ctx)
		if nil != err {
			return nil, fmt.Errorf("failed to create googleCloud client: %s", err)
		}
	}

	return &googleCloud{
		ctx:    ctx,
		lang:   lang,
		client: client,
	}, err
}
