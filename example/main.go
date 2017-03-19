package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ccyun/octopus"
)

func main() {
	file := filepath.Dir(os.Args[0])
	path, _ := filepath.Abs(file)
	log.Println(path)
	o := new(octopus.Octopus)
	o.Run()
}
