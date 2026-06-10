package liveobjects

func StringKeys(input map[any]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any)
	for k, v := range input {
		strKey, ok := k.(string)
		if !ok {
			continue
		}
		out[strKey] = v
	}
	return out
}