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
