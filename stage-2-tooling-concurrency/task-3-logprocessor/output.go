package main

import (
	"encoding/json"
	"io"
)

func writeOutput(w io.Writer, res *Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}