package translator

import (
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

type googleCloudAPI interface {
	Translate(context.Context, []string, language.Tag, *gcp.Options) ([]gcp.Translation, error)
}

type googleCloud struct {
	sourceFileName string
	ctx            context.Context
	client         googleCloudAPI
	lang           language.Tag
	output         *os.File
	content        string
}

func (g *googleCloud) Initialise(sourceFileName string) error {
	var err error
	g.sourceFileName = sourceFileName

	g.output, err = os.Create(outputFileName)
	if nil != err {
		return fmt.Errorf("fail to create output file %s with error: %s", outputFileName, err)
	}

	content, err := ioutil.ReadFile(g.sourceFileName)
	if nil != err {
		return fmt.Errorf("read file %s with error: %s", g.sourceFileName, err)
	}

	var sb strings.Builder
	for _, b := range content {
		sb.WriteByte(b)
	}
	g.content = sb.String()

	return nil
}

func (g *googleCloud) Translate() error {
	sleepTimeMillisecond := float64(1000) / gcpAPIRateLimit

	startIndex := 0
	var bs string
	for startIndex < len(g.content)-1 {
		bs, startIndex = translatedWords(g.content, startIndex)

		// Translate
		translated, err := g.client.Translate(g.ctx, []string{bs}, g.lang, &gcp.Options{
			Format: gcp.Text,
		})
		if nil != err {
			return fmt.Errorf("api query error: %v", err)
		}

		time.Sleep(time.Duration(sleepTimeMillisecond) * time.Millisecond)

		err = writeToFile(translated, err, g)
		if err != nil {
			return fmt.Errorf("write error: %s\n", err)
		}
	}

	fmt.Printf("total translated characters: %d\n", len(g.content))

	return nil
}

func writeToFile(translated []gcp.Translation, err error, g *googleCloud) error {
	for _, s := range translated {
		_, err = g.output.WriteString(s.Text)
		if nil != err {
			return fmt.Errorf("write to file %s with error: %s", outputFileName, err)
		}
	}
	err = g.output.Sync()
	if nil != err {
		return fmt.Errorf("sync file %s with error: %s", outputFileName, err)
	}

	return nil
}

func translatedWords(content string, startIndex int) (string, int) {
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

	offsetIndex := strings.LastIndex(content[startIndex:index+1], ".")

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
