package batcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (c *CHBatch) Insert(ctx context.Context, b *Batcher) error {
	insertAnyStat(ctx, "devices_traffics_5s", c.DeviceTraffics, b)
	insertAnyStat(ctx, "devices_domains_5s", c.DeviceDomains, b)
	insertAnyStat(ctx, "devices_countries_5s", c.DeviceCountries, b)
	insertAnyStat(ctx, "devices_protos_5s", c.DeviceProtos, b)

	return nil
}

func insertAnyStat[T DeviceStatLike](ctx context.Context, table_name string, records []T, b *Batcher) error {
	if len(records) == 0 {
		return nil
	}

	per_device_id := aggregatePerDeviceID(records)
	
	for device_id, records_per_device := range per_device_id {
		insertStatPerDevice(ctx, records_per_device, b, device_id)
	}
	return nil
}

func insertStatPerDevice[T DeviceStatLike](ctx context.Context, records []T, b *Batcher, device_id uint64) error {
	if len(records) == 0 {
		return nil
	}

	for _, rec := range records {
		switch r := any(rec).(type) {
		case DeviceTraffic:
			// increment numeric columns on conflict (device_id, time)
			q := `INSERT INTO devices_traffics_5s (time, device_id, up_bytes, req_count)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (device_id, time) DO UPDATE
				  SET up_bytes = devices_traffics_5s.up_bytes + EXCLUDED.up_bytes,
					  req_count = devices_traffics_5s.req_count + EXCLUDED.req_count`
			if err := b.PGDB.WithContext(ctx).Exec(q, r.Bucket, uint64(r.DeviceID), r.UpBytes, r.Requests).Error; err != nil {
				return err
			}
		case DeviceDomain:
			// increment requests for same domain (device_id, time, domain)
			q := `INSERT INTO devices_domains_5s (time, device_id, domain, requests)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (device_id, time, domain) DO UPDATE
				  SET requests = devices_domains_5s.requests + EXCLUDED.requests`
			if err := b.PGDB.WithContext(ctx).Exec(q, r.Bucket, uint64(r.DeviceID), string(r.Domain), r.Requests).Error; err != nil {
				return err
			}
		case DeviceCountry:
			// merge arrays (companies, countries) and sum requests on conflict (device_id, time)
			// coalesce is used to handle NULL arrays
			q := `INSERT INTO devices_countries_5s (time, device_id, companies, countries, requests)
				  VALUES ($1, $2, $3, $4, $5)
				  ON CONFLICT (device_id, time) DO UPDATE
				  SET companies = (
						  SELECT ARRAY(SELECT DISTINCT x FROM unnest(coalesce(devices_countries_5s.companies, '{}') || EXCLUDED.companies) AS x)
					  ),
					  countries = (
						  SELECT ARRAY(SELECT DISTINCT x FROM unnest(coalesce(devices_countries_5s.countries, '{}') || EXCLUDED.countries) AS x)
					  ),
					  requests = devices_countries_5s.requests + EXCLUDED.requests`
			if err := b.PGDB.WithContext(ctx).Exec(q, r.Bucket, uint64(r.DeviceID), pq.StringArray(r.Company), pq.StringArray(r.Country), r.Requests).Error; err != nil {
				return err
			}
		case DeviceProto:
			// increment requests for same proto (device_id, time, proto)
			q := `INSERT INTO devices_protos_5s (time, device_id, proto, requests)
				  VALUES ($1, $2, $3, $4)
				  ON CONFLICT (device_id, time, proto) DO UPDATE
				  SET requests = devices_protos_5s.requests + EXCLUDED.requests`
			if err := b.PGDB.WithContext(ctx).Exec(q, r.Bucket, uint64(r.DeviceID), string(r.Proto), r.Requests).Error; err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported record type: %T", records[0])
		}
	}

	return nil
}

func (b *Batcher) checkAlreadyHadTimeBatches(ctx context.Context, times []time.Time, table_name string, device_id uint64) ([]time.Time, []time.Time, error) {
	if len(times) == 0 {
		return nil, nil, nil
	}

	// Для больших слайсов разбиваем на чанки, чтобы не строить гигантский IN(...)
	const chunkSize = 5000

	existing := make(map[int64]struct{}, len(times))

	for i := 0; i < len(times); i += chunkSize {
		end := i + chunkSize
		if end > len(times) {
			end = len(times)
		}
		chunk := times[i:end]

		// Build placeholders and args for this chunk
		args := make([]interface{}, 0, len(chunk)+1)
		args = append(args, device_id)

		placeholders := make([]string, len(chunk))
		for j, t := range chunk {
			placeholders[j] = "?"
			args = append(args, t)
		}

		query := fmt.Sprintf("SELECT time FROM %s WHERE device_id = ? AND time IN (%s)", table_name, strings.Join(placeholders, ","))

		rows, err := b.PGDB.WithContext(ctx).Raw(query, args...).Rows()
		if err != nil {
			return nil, nil, err
		}

		for rows.Next() {
			var t time.Time
			if err := rows.Scan(&t); err != nil {
				rows.Close()
				return nil, nil, err
			}
			existing[t.UnixNano()] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, nil, err
		}
		rows.Close()
	}

	have := make([]time.Time, 0, len(existing))
	notHave := make([]time.Time, 0, len(times)-len(existing))
	for _, t := range times {
		if _, ok := existing[t.UnixNano()]; ok {
			have = append(have, t)
		} else {
			notHave = append(notHave, t)
		}
	}

	return have, notHave, nil
}