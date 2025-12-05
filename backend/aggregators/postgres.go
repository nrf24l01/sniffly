package aggregators

import (
	"time"

	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"gorm.io/gorm"
)

func loadFromPostgres(db *gorm.DB, times_to_load []time.Time) ([]analyzerModels.DeviceTraffic5s, error) {
	results := make([]analyzerModels.DeviceTraffic5s, 0)

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		batchResults := make([]analyzerModels.DeviceTraffic5s, 0)
		if err := db.Where("bucket >= ? AND bucket < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}