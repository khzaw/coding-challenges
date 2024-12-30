package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

// -c  - outputs the number of bytes in a file
// -l  - outputs the number of lines in a file
// -w  - outputs the number of words in a file

func main() {
	bytesOpt := flag.Bool("c", false, "number of bytes in a file")
	lineOpt := flag.Bool("l", false, "number of lines in a file")
	wordOpt := flag.Bool("w", false, "number of words in a file")
	flag.Parse()

	if filename := flag.Arg(0); filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("error opening file: err:", err)
			os.Exit(1)
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			fmt.Println("error getting file stat: err:", err)
			os.Exit(1)
		}

		if *bytesOpt {
			fmt.Printf("%d %s\n", fi.Size(), fi.Name())
		}

		if *lineOpt {
			lines := 0
			buf := bufio.NewScanner(f)
			for buf.Scan() {
				buf.Text()
				lines += 1
			}
			fmt.Printf("%d %s\n", lines, fi.Name())
		}

		if *wordOpt {

		}
	}
}
