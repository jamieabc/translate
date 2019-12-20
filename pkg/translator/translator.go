package translator

import "fmt"

const (
	GCP = iota
)

type Translator interface {
	Translate(string) error
}

func NewTranslator(platform int) (Translator, error) {
	switch platform {
	case GCP:
		return &googleCloud{}, nil
	default:
		return nil, fmt.Errorf("wrong platform")
	}
}
