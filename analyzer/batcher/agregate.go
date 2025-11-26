package batcher

import "time"


func aggregatePerTime[T DeviceStatLike](stats []T) (map[time.Time][]interface{}, []time.Time) {
	var result = make(map[time.Time][]interface{})
	
	for _, dt := range stats {
		result[dt.Bucket] = append(result[dt.Bucket], dt)
	}
	
	keys := make([]time.Time, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}

	return result, keys
}

func aggregatePerDeviceID[T DeviceStatLike](records []T) map[uint64][]T {
	// Input agregation per date output per device ID
	result := make(map[uint64][]T)

	for _, v := range records {
		result[device_id] = append(result[device_id], v)
	}

	return result
}

func aggregatePerDateToPedDatePerDeviceID(per_date map[time.Time][]interface{}) map[time.Time]map[uint64][]interface{} {
	// Input agregation per date output per date per device ID
	result := make(map[time.Time]map[uint64][]interface{})

	for date, records := range per_date {
		if _, exists := result[date]; !exists {
			result[date] = make(map[uint64][]interface{})
		}
		for _, record := range records {
			switch v := record.(type) {
			case *DeviceTraffic:
				device_id := v.DeviceID
				result[date][device_id] = append(result[date][device_id], v)
			case *DeviceDomain:
				device_id := v.DeviceID
				result[date][device_id] = append(result[date][device_id], v)
			case *DeviceCountry:
				device_id := v.DeviceID
				result[date][device_id] = append(result[date][device_id], v)
			case *DeviceProto:
				device_id := v.DeviceID
				result[date][device_id] = append(result[date][device_id], v)
			}
		}
	}

	return result
}
