package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

type Config struct {
	flagSize  bool
	flagLines bool
	flagWords bool
	flagChars bool
	f         *os.File
	size      int
	lines     int
	words     int
	chars     int
}

func (c *Config) count(buf *bufio.Scanner) int {
	count := 0
	for buf.Scan() {
		buf.Text()
		count += 1
	}
	return count
}

func (c *Config) ScanToken(splitFn bufio.SplitFunc) int {
	var r io.Reader
	if filename := flag.Arg(0); filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("error opening file: ", err)
			os.Exit(1)
		}
		r = f
		defer f.Close()
	} else {
		r = os.Stdin
	}

	buf := bufio.NewScanner(r)
	buf.Split(splitFn)
	return c.count(buf)
}

func parseFlags() *Config {
	// -c  - outputs the number of bytes in a file
	// -l  - outputs the number of lines in a file
	// -w  - outputs the number of words in a file
	// -m  - outputs the number of characters in a file
	c := flag.Bool("c", false, "number of bytes in a file")
	l := flag.Bool("l", false, "number of lines in a file")
	w := flag.Bool("w", false, "number of words in a file")
	m := flag.Bool("m", false, "number of chars in a file")
	flag.Parse()

	return &Config{flagSize: *c, flagLines: *l, flagWords: *w, flagChars: *m}
}

func main() {

	cfg := parseFlags()
	if !cfg.flagSize && !cfg.flagLines && !cfg.flagWords && !cfg.flagChars {
		cfg.flagSize = true
		cfg.flagLines = true
		cfg.flagWords = true
	}

	var outputStr string

	if cfg.flagLines {
		outputStr += fmt.Sprintf("     %d", cfg.ScanToken(bufio.ScanLines))
	}

	if cfg.flagWords {
		outputStr += fmt.Sprintf("     %d", cfg.ScanToken(bufio.ScanWords))
	}

	if cfg.flagSize {
		outputStr += fmt.Sprintf("     %d", cfg.ScanToken(bufio.ScanRunes))
	}

	if cfg.flagChars {
		outputStr += fmt.Sprintf("     %d", cfg.ScanToken(bufio.ScanBytes))
	}

	fmt.Println(outputStr)
}
