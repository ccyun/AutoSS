package octopus

import (
	"log"
	"testing"
)

func Test_deqr(t *testing.T) {
	o := new(Octopus)
	o.deqr()
	log.Println(o.Config)
}
