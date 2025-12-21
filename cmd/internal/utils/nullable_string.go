package utils

import "encoding/json"

type NullableString struct {
	Value string
}

func (ns *NullableString) UnmarshalJSON(data []byte) error {
	// false or null → empty string
	if string(data) == "false" || string(data) == "null" {
		ns.Value = ""
		return nil
	}

	// otherwise must be string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ns.Value = s
	return nil
}
