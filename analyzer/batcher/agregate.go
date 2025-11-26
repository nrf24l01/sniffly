package batcher

import (
	"time"

	"github.com/google/uuid"
)


func aggregatePerTime[T DeviceStatLike](stats []T) (map[time.Time][]interface{}, []time.Time) {
	var result = make(map[time.Time][]interface{})
	
	for _, dt := range stats {
		result[dt.GetBucket()] = append(result[dt.GetBucket()], dt)
	}
	
	keys := make([]time.Time, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}

	return result, keys
}

func aggregatePerDeviceID[T DeviceStatLike](records []T) map[uuid.UUID][]T {
	// Input agregation per date output per device ID
	result := make(map[uuid.UUID][]T)

	for _, v := range records {
		id := v.GetDeviceID()
		result[id] = append(result[id], v)
	}

	return result
}