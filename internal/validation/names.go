package validation

import "unicode/utf8"

func WorkspaceID(value string) bool {
	if len(value) < 2 || len(value) > 48 || !utf8.ValidString(value) {
		return false
	}
	for index, item := range value {
		if item >= 'a' && item <= 'z' || item >= '0' && item <= '9' {
			continue
		}
		if item == '-' && index > 0 && index < len(value)-1 {
			continue
		}
		return false
	}
	return value[0] >= 'a' && value[0] <= 'z'
}

func AppKey(value string) bool {
	if len(value) < 2 || len(value) > 64 || !utf8.ValidString(value) {
		return false
	}
	for _, item := range value {
		if item >= 'a' && item <= 'z' ||
			item >= 'A' && item <= 'Z' ||
			item >= '0' && item <= '9' ||
			item == '_' {
			continue
		}
		return false
	}
	return true
}
