package translator_test

import (
	"context"
	"os"
	"testing"

	gcp "cloud.google.com/go/translate"
	"golang.org/x/text/language"

	"github.com/jamieabc/translate/pkg/translator"
	"github.com/stretchr/testify/assert"
)

const (
	outputFilename = "out.txt"
)

type MockClient struct {
	text  []string
	times int
}

func (c *MockClient) Translate(ctx context.Context, inputs []string, target language.Tag, opts *gcp.Options) ([]gcp.Translation, error) {
	if c.times == 0 {
		c.text = inputs
		c.times++
	}

	return []gcp.Translation{gcp.Translation{
		Text:   "",
		Source: language.Tag{},
		Model:  "",
	}}, nil
}

func removeFile() {
	_ = os.Remove(outputFilename)
}

func TestGoogleCloud_Translate(t *testing.T) {
	mock := &MockClient{}
	g, err := translator.NewTranslator(translator.GCP, mock)
	assert.Nil(t, err, "wrong error")

	err = g.Initialise("fixtures/test.txt")
	defer removeFile()

	assert.Nil(t, err, "wrong error")

	err = g.Translate()
	assert.Nil(t, err, "wrong error")

	expected1 := "This is a test file to test. This has no meaning, not at all. It is just for filling to limit of 123. And this is the last line to exceed limitation of 128 characters per api request.\n\nThis line should be included in test.\n"
	assert.Equal(t, expected1, mock.text[0], "wrong text")
}
