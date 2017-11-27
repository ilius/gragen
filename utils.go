package main

func tagHasFlag(tagParts []string, flag string) bool {
	for _, part := range tagParts {
		if part == flag {
			return true
		}
	}
	return false
}
