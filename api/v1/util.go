package v1

func containsString(array []string, input string) bool {
	for _, item := range array {
		if item == input {
			return true
		}
	}
	return false
}

func removeString(array []string, input string) (result []string) {
	for _, item := range array {
		if item == input {
			continue
		}
		result = append(result, item)
	}
	return result
}
