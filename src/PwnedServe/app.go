package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Range struct {
	min int64
	max int64
}

var HashFile *os.File
var HashCount int64
var HashCache map[int32]Range

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fileName := homeDir + "/Downloads/Pwned/pwned-passwords-sha1-ordered-by-hash-v7.bin"

	file, err := os.Open(fileName)
	if err != nil {
		panic("Cannot open input file")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		panic("Cannot stat input file")
	}

	HashCache = make(map[int32]Range)
	HashCount = info.Size() / 20
	HashFile = file

	localAddress := ":8080"
	fmt.Println("Listening at " + localAddress)
	http.HandleFunc("/range/", handler)
	log.Fatal(http.ListenAndServe(localAddress, nil))

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

func handler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	lastPart := pathParts[len(pathParts)-1]

	if len(lastPart) == 5 {

		bytes, err := hex.DecodeString("0" + lastPart)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		prefix := int32(bytes[0])<<16 | int32(bytes[1])<<8 | int32(bytes[2])
		indexRange := findMatchRange(prefix)

		var output strings.Builder
		for index := indexRange.min; index <= indexRange.max; index++ {
			buffer := readHashAt(index)
			output.WriteString(hex.EncodeToString(buffer))
			output.WriteByte('\n')
		}

		w.Write([]byte(output.String()))

	} else if len(lastPart) == 40 {

		bytes, err := hex.DecodeString(lastPart)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		prefix := int32(bytes[0])<<16 | int32(bytes[1])<<8 | int32(bytes[2])
		indexRange := findMatchRange(prefix)

		min := indexRange.min
		max := indexRange.max

		isOk := true
		var currBuffer []byte
		for {
			if min > max {
				isOk = false
				break
			}

			pivot := (min + max) / 2
			currBuffer = readHashAt(pivot)

			shouldRepeat := false
			for i := 2; i < 20; i++ {
				if bytes[i] < currBuffer[i] {
					max = pivot - 1
					shouldRepeat = true
					break
				} else if bytes[i] > currBuffer[i] {
					min = pivot + 1
					shouldRepeat = true
					break
				}
			}

			if !shouldRepeat {
				break
			}
		}

		if isOk {
			w.Write([]byte(hex.EncodeToString(currBuffer)))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

func findMatchRange(prefix int32) Range {
	if indexRange, found := HashCache[prefix]; found {
		return indexRange
	}

	index := findMatch(prefix, 0, HashCount-1)

	indexMax := index
	for indexMax < HashCount-1 {
		nextPrefix := readHashPrefixAt(indexMax + 1)
		if nextPrefix != prefix {
			break
		}
		indexMax++
	}

	indexMin := index
	for indexMin > 0 {
		prevPrefix := readHashPrefixAt(indexMin - 1)
		if prevPrefix != prefix {
			break
		}
		indexMin--
	}

	indexRange := Range{min: indexMin, max: indexMax}
	HashCache[prefix] = indexRange
	return indexRange
}

func findMatch(prefix int32, minIndex int64, maxIndex int64) int64 {
	if minIndex > maxIndex {
		return -1
	}

	pivot := (minIndex + maxIndex) / 2
	currPrefix := readHashPrefixAt(pivot)
	if currPrefix == prefix {
		return pivot
	} else if prefix < currPrefix {
		return findMatch(prefix, minIndex, pivot-1)
	} else {
		return findMatch(prefix, pivot+1, maxIndex)
	}
}

func readHashPrefixAt(index int64) int32 {
	buffer := readHashAt(index)
	prefix := int32(buffer[0])<<12 | int32(buffer[1])<<4 | int32(buffer[2])>>4
	return prefix
}

func readHashAt(index int64) []byte {
	offset := index * 20
	buffer := make([]byte, 20)
	HashFile.ReadAt(buffer, offset)
	return buffer
}
