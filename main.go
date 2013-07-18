package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("run sql2json /path/to/file.sql > /path/to/export.json")
		return
	}

	filename := os.Args[1]

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	data := Convert(f)
	fmt.Println(string(data))
}
