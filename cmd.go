package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"roob.re/omemo-wget/aesgcm"
	"strings"
)

func main() {
	outfile := flag.String("o", "", "out file. Use '-' for stdout. Defaults to guess from input uri/path")
	flag.Parse()

	if flag.NArg() < 1 {
		stderrExit(errors.New(fmt.Sprintf("Usage: %s <path/to/file|uri#hash> [hash] [-o out]\n", os.Args[0])), 1)
		return
	}

	uri := flag.Args()[0]

	var hash string
	parts := strings.Split(uri, "#")
	path := parts[0]
	if len(parts) >= 2 {
		hash = parts[1]
	} else if flag.NArg() >= 2 {
		hash = flag.Args()[1]
	} else {
		stderrExit(errors.New("hash must be either included in the url (after the # character) or provided as a second argument"), 2)
		return
	}

	in, err := open(path)
	if err != nil {
		stderrExit(err, 3)
		return
	}

	fileContents, err := ioutil.ReadAll(in)
	if err != nil {
		stderrExit(err, 4)
		return
	}
	_ = in.Close()

	decryptedContents, err := aesgcm.Decrypt(fileContents, hash)
	if err != nil {
		panic(err)
	}

	var out io.WriteCloser
	switch *outfile {
	case "-":
		out = os.Stdout
	case "":
		// Generate a suitable name
		basename := filepath.Base(path)
		ext := filepath.Ext(basename)

		for _, err := os.Stat(basename); err == nil; _, err = os.Stat(basename) {
			basename = strings.Replace(basename, ext, "_decrypted"+ext, 1)
		}
		*outfile = basename
		fallthrough
	default:
		f, err := os.Create(*outfile)
		if err != nil {
			stderrExit(errors.New("error creating output file: "+err.Error()), 6)
			return
		}
		out = f
	}

	_, err = out.Write(decryptedContents)
	if err != nil {
		stderrExit(errors.New("error writing to output file: "+err.Error()), 6)
		return
	}
	_ = out.Close()
}

func open(uri string) (io.ReadCloser, error) {
	switch true {
	case strings.HasPrefix(uri, "aesgcm"):
		uri = strings.Replace(uri, "aesgcm", "https", 1)
		fallthrough
	case strings.HasPrefix(uri, "https"):
		resp, err := http.Get(uri)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error fetching '%s': %s", uri, err.Error()))
		}

		return resp.Body, nil
	default:
		file, err := os.Open(uri)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("could not open file '%s': %s", uri, err.Error()))
		}

		return file, nil
	}
}

func stderrExit(e error, code int) {
	_, _ = fmt.Fprint(os.Stderr, e.Error()+"\n")
	os.Exit(code)
}
