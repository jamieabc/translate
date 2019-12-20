package translator

import "fmt"

const (
	GCP = iota
)

type Translator interface {
	Initialise(string) error
	Translate() error
}

func NewTranslator(platform int) (Translator, error) {
	switch platform {
	case GCP:
		return newGCP()
	default:
		return nil, fmt.Errorf("wrong platform")
	}
}
