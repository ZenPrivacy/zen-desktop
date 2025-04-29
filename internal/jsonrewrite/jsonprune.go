package jsonrewrite

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func ModifyJSONResponse(resp *http.Response, propsToRemove string, requiredProps string) error {
	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var jsonData map[string]any
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		return nil // ignore if not valid JSON
	}

	propsToRemovePaths := parsePropPaths(propsToRemove)
	requiredPropsPaths := parsePropPaths(requiredProps)

	pruneJSON(jsonData, propsToRemovePaths, requiredPropsPaths)

	newBody, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	resp.Body = io.NopCloser(bytes.NewReader(newBody))
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
	return nil
}

func parsePropPaths(input string) [][]string {
	if input == "" {
		return nil
	}
	var result [][]string
	for _, path := range strings.Fields(input) {
		result = append(result, strings.Split(path, "."))
	}
	return result
}

func hasRequiredProps(data map[string]any, paths [][]string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, path := range paths {
		if matchesPath(data, path) {
			return true
		}
	}
	return false
}

func matchesPath(obj any, path []string) bool {
	if obj == nil || len(path) == 0 {
		return true
	}

	current := path[0]
	rest := path[1:]

	switch current {
	case "[]":
		arr, ok := obj.([]any)
		if !ok {
			return false
		}
		for _, item := range arr {
			if matchesPath(item, rest) {
				return true
			}
		}
		return false
	}

	switch o := obj.(type) {
	case map[string]any:
		val, ok := o[current]
		if !ok {
			return false
		}
		return matchesPath(val, rest)
	case []any:
		for _, item := range o {
			if matchesPath(item, path) {
				return true
			}
		}
	}
	return false
}

func prunePath(obj any, path []string) {
	if obj == nil || len(path) == 0 {
		return
	}

	current := path[0]
	rest := path[1:]

	switch current {
	case "[]":
		arr, ok := obj.([]any)
		if !ok {
			return
		}
		for _, item := range arr {
			prunePath(item, rest)
		}
		return
	}

	switch o := obj.(type) {
	case map[string]any:
		if val, ok := o[current]; ok {
			if len(rest) == 0 {
				delete(o, current)
			} else {
				prunePath(val, rest)
			}
		}
	case []any:
		for _, item := range o {
			prunePath(item, path)
		}
	}
}

func pruneJSON(data map[string]any, propsToRemove [][]string, requiredProps [][]string) {
	if !hasRequiredProps(data, requiredProps) {
		return
	}
	for _, path := range propsToRemove {
		prunePath(data, path)
	}
}
