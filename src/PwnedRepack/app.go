package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	inFileName := homeDir + "/Downloads/Pwned/pwned-passwords-sha1-ordered-by-hash-v7.txt"
	outFileName := homeDir + "/Downloads/Pwned/pwned-passwords-sha1-ordered-by-hash-v7.bin"

	inFile, err := os.Open(inFileName)
	if err != nil {
		panic("Cannot open input file")
	}
	defer inFile.Close()

	outFile, err := os.Create(outFileName)
	if err != nil {
		panic("Cannot create output file")
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	lineNumber := 0
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		lineNumber++
		if lineNumber%1000000 == 0 {
			if lineNumber%10000000 == 0 {
				fmt.Print(":")
			} else {
				fmt.Print(".")
			}
		}

		line := scanner.Text()
		sha1Hex := strings.Split(line, ":")[0]
		sha1Bytes, err := hex.DecodeString(sha1Hex)
		if err != nil {
			panic(err)
		}

		n, err := writer.Write(sha1Bytes)
		if n != 20 {
			panic("Invalid count")
		}
		if err != nil {
			panic(err)
		}
	}
	fmt.Println()
	fmt.Printf("Done (%d lines, %d bytes).\n", lineNumber, lineNumber*20)

	writer.Flush()
	os.Exit(0)

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
