package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func gunzipBytes(gzBytes []byte) ([]byte, error) {
	gzReader := bytes.NewReader(gzBytes)
	reader, err := gzip.NewReader(gzReader)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func tagHasFlag(tagParts []string, flag string) bool {
	for _, part := range tagParts {
		if part == flag {
			return true
		}
	}
	return false
}
