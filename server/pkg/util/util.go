package util

import "time"

func ChunkStringSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func IsInLastWeek(t time.Time) bool {
	// Get the current time
	now := time.Now()

	// Calculate the duration between now and the provided time
	duration := now.Sub(t)

	// Check if the duration is within the last week (7 days)
	return duration.Seconds() <= 7*24*60*60
}

func CombineMaps(map1, map2 map[string]int) map[string]int {
	result := make(map[string]int)

	// Copy the values from the first map
	for key, value := range map1 {
		result[key] = value
	}

	// Add or update values from the second map
	for key, value := range map2 {
		if existingValue, ok := result[key]; ok {
			// Key already exists, add the values
			result[key] = existingValue + value
		} else {
			// Key doesn't exist, add a new entry
			result[key] = value
		}
	}

	return result
}
