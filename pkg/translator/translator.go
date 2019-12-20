package translator

import "fmt"

const (
	GCP = iota
)

type Translator interface {
	Initialise() error
	Translate() error
}

func NewTranslator(platform int, sourceFileName string) (Translator, error) {
	switch platform {
	case GCP:
		return newGCP(sourceFileName), nil
	default:
		return nil, fmt.Errorf("wrong platform")
	}
}
