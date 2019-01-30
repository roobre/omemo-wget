package main

import (
	"github.com/pkg/errors"
	"io"
	"net/http"
	"roob.re/omemo-wget/aesgcm"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <path/to/file|https://url#hash> [hash]\n", os.Args[0])
		os.Exit(1)
	}

	str := os.Args[1]
	var in io.ReadCloser
	var hash string
	var filename string

	parts := strings.Split(str, "/")
	filename = parts[len(parts)-1]

	if strings.HasPrefix(str, "aesgcm://") {
		str = strings.Replace(str, "aesgcm://", "https://", 1)
	}

	if strings.HasPrefix(str, "https://") {
		resp, err := http.Get(str)
		if err != nil {
			stderrExit(err, 2)
		}

		in = resp.Body
		parts := strings.Split(filename, "#")
		if len(parts) < 2 {
			stderrExit(errors.New("Malformed aesgcm url"), 5)
		}
		hash = parts[len(parts)-1]
		filename = parts[len(parts) - 2]
	} else {
		if len(os.Args) < 3 {
			_, _ = fmt.Fprintf(os.Stderr, "Hash is mandatory if not included in the url\n\nUsage: %s <path/to/file|https://url#hash> [hash]\n", os.Args[0])
			os.Exit(3)
		}
		hash = os.Args[2]

		var err error
		in, err = os.Open(str)
		if err != nil {
			stderrExit(err, 2)
		}
	}

	outfile := filename
	pos := strings.IndexAny(filename, ".")
	if pos != -1 {
		outfile = filename[:pos] + "_decrypted" + filename[pos:]
	} else {
		outfile += "_decrypted"
	}

	fileContents, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	in.Close()

	decryptedContents, err := aesgcm.Decrypt(fileContents, hash)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(outfile)
	if err != nil {
		stderrExit(err, 5)
	}
	out.Write(decryptedContents)
	out.Close()

}

func stderrExit(e error, code int) {
	_, _ = fmt.Fprint(os.Stderr, e)
	os.Exit(code)
}
