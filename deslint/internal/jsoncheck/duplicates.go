// Package jsoncheck provides strict JSON checks shared by contract inputs.
package jsoncheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// DuplicateKeys returns sorted JSON paths that contain duplicate object keys.
func DuplicateKeys(contents []byte) ([]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	duplicates := make([]string, 0)
	if err := decodeValue(decoder, "$", &duplicates); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, fmt.Errorf("multiple JSON values")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return nil, fmt.Errorf("multiple JSON values")
	}
	sort.Strings(duplicates)
	return duplicates, nil
}

func decodeValue(decoder *json.Decoder, path string, duplicates *[]string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, tokenErr := decoder.Token()
			if tokenErr != nil {
				return tokenErr
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key at %s is not a string", path)
			}
			childPath := path + "." + key
			if _, exists := seen[key]; exists {
				*duplicates = append(*duplicates, childPath)
			}
			seen[key] = struct{}{}
			if decodeErr := decodeValue(decoder, childPath, duplicates); decodeErr != nil {
				return decodeErr
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		index := 0
		for decoder.More() {
			if decodeErr := decodeValue(decoder, fmt.Sprintf("%s[%d]", path, index), duplicates); decodeErr != nil {
				return decodeErr
			}
			index++
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected delimiter %q", delimiter)
	}
}
