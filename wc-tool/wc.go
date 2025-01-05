package main

import (
	"bufio"
	"flag"
	"fmt"
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

func (c *Config) ScanToken(filename string, splitFn bufio.SplitFunc) int {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file: ", err)
		os.Exit(1)
	}
	defer f.Close()

	buf := bufio.NewScanner(f)
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

	if filename := flag.Arg(0); filename != "" {
		if cfg.flagLines {
			outputStr += fmt.Sprintf("     %d", cfg.ScanToken(filename, bufio.ScanLines))
		}

		if cfg.flagWords {
			outputStr += fmt.Sprintf("     %d", cfg.ScanToken(filename, bufio.ScanWords))
		}

		if cfg.flagSize {
			outputStr += fmt.Sprintf("     %d", cfg.ScanToken(filename, bufio.ScanRunes))
		}

		if cfg.flagChars {
			outputStr += fmt.Sprintf("     %d", cfg.ScanToken(filename, bufio.ScanBytes))
		}

		fmt.Println(outputStr, filename)
	}
}
