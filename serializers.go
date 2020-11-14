package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"sort"
	"strings"
)

type serializer interface {
	serialize(stringOnlyJSON) string
}

type jsonSerializer struct {
}

func (js jsonSerializer) serialize(soj stringOnlyJSON) string {
	serialized, _ := json.Marshal(soj)
	return string(serialized)
}

type csvSerializer struct {
}

func (cs csvSerializer) serialize(soj stringOnlyJSON) string {
	keys := make([]string, len(soj))
	keyIdx := 0
	for k := range soj {
		keys[keyIdx] = k
		keyIdx++
	}

	sort.Strings(keys)

	values := make([]string, len(keys))
	valueIdx := 0
	for _, k := range keys {
		values[valueIdx] = soj[k]
		valueIdx++
	}

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	w.Write(values)
	w.Flush()

	s := b.String()
	return strings.TrimSpace(s)
}
