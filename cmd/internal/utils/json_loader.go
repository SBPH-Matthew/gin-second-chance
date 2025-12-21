package utils

import (
	"encoding/json"
	"os"
)

func LoadJSON[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result []T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
