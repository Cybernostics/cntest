package random

import (
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
)

var seed = time.Now().UTC().UnixNano()
var nameMaker = namegenerator.NewNameGenerator(seed)

// Name generates a random name
func Name() string {
	return strings.Replace(nameMaker.Generate(), "-", "", 1)
}
