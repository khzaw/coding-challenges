package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// -c  - outputs the number of bytes in a file
// -l  - outputs the number of lines in a file
// -w  - outputs the number of words in a file
// -m  - outputs the number of characters in a file

func main() {
	bytesOpt := flag.Bool("c", false, "number of bytes in a file")
	lineOpt := flag.Bool("l", false, "number of lines in a file")
	wordOpt := flag.Bool("w", false, "number of words in a file")
	charOpt := flag.Bool("m", false, "number of characters in a file")
	flag.Parse()

	if filename := flag.Arg(0); filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("error opening file: err:", err)
			os.Exit(1)
		}
		defer f.Close()

		path, _ := filepath.Abs(filename)

		fi, err := f.Stat()
		if err != nil {
			fmt.Println("error getting file stat: err:", err)
			os.Exit(1)
		}

		if *bytesOpt {
			fmt.Printf("%d %s\n", fi.Size(), path)
		}

		if *lineOpt {
			lines := 0
			buf := bufio.NewScanner(f)
			for buf.Scan() {
				buf.Text()
				lines += 1
			}
			fmt.Printf("%d %s\n", lines, path)
		}

		if *wordOpt {
			buf := bufio.NewScanner(f)
			buf.Split(bufio.ScanWords)
			words := 0
			for buf.Scan() {
				buf.Text()
				words += 1
			}
			fmt.Printf("%d %s\n", words, path)
		}

		if *charOpt {
			buf := bufio.NewScanner(f)
			buf.Split(bufio.ScanRunes)
			chr := 0
			for buf.Scan() {
				buf.Text()
				chr += 1
			}
			fmt.Printf("%d %s\n", chr, path)
		}
	}
}
