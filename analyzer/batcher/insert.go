package batcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (c *CHBatch) Insert(ctx context.Context, b *Batcher) error {
	// Use typed insert helper (fixed table names inside) to avoid dynamic SQL identifiers
	insertAnyStat(ctx, c.DeviceTraffics, b)
	insertAnyStat(ctx, c.DeviceDomains, b)
	insertAnyStat(ctx, c.DeviceCountries, b)
	insertAnyStat(ctx, c.DeviceProtos, b)

	return nil
}

func insertAnyStat[T DeviceStatLike](ctx context.Context, records []T, b *Batcher) error {
	if len(records) == 0 {
		return nil
	}

	per_device_id := aggregatePerDeviceID(records)

	for device_id, records_per_device := range per_device_id {
		insertStatPerDevice(ctx, records_per_device, b, device_id)
	}
	return nil
}

func insertStatPerDevice[T DeviceStatLike](ctx context.Context, records []T, b *Batcher, device_id uuid.UUID) error {
	if len(records) == 0 {
		return nil
	}

	var traffics []DeviceTraffic
	var domains []DeviceDomain
	var countries []DeviceCountry
	var protos []DeviceProto

	for _, rec := range records {
		switch r := any(rec).(type) {
		case DeviceTraffic:
			traffics = append(traffics, r)
		case DeviceDomain:
			domains = append(domains, r)
		case DeviceCountry:
			countries = append(countries, r)
		case DeviceProto:
			protos = append(protos, r)
		default:
			return fmt.Errorf("unsupported record type: %T", rec)
		}
	}

	exec := func(q string, args ...interface{}) error {
		if q == "" {
			return nil
		}
		if err := b.PGDB.WithContext(ctx).Exec(q, args...).Error; err != nil {
			return err
		}
		return nil
	}

	// Batch DeviceTraffic
	if len(traffics) > 0 {
		cols := "bucket,device_id,up_bytes,req_count"
		var vals []string
		var args []interface{}
		for i, r := range traffics {
			base := i * 4
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4))
			args = append(args, r.Bucket, r.DeviceID, r.UpBytes, r.Requests)
		}
		q := fmt.Sprintf(`INSERT INTO devices_traffics_5s (%s) VALUES %s
			ON CONFLICT (device_id, bucket) DO UPDATE
			SET up_bytes = devices_traffics_5s.up_bytes + EXCLUDED.up_bytes,
				req_count = devices_traffics_5s.req_count + EXCLUDED.req_count`, cols, strings.Join(vals, ","))
		if err := exec(q, args...); err != nil {
			return err
		}
	}

	// Batch DeviceDomain
	if len(domains) > 0 {
		cols := "bucket,device_id,domain,requests"
		var vals []string
		var args []interface{}
		for i, r := range domains {
			base := i * 4
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4))
			args = append(args, r.Bucket, r.DeviceID, string(r.Domain), r.Requests)
		}
			q := fmt.Sprintf(`INSERT INTO devices_domains_5s (%s) VALUES %s
				ON CONFLICT (device_id, bucket) DO UPDATE
				SET domain = (
					SELECT jsonb_object_agg(k, to_jsonb(sum_v)) FROM (
						SELECT k, sum(v::bigint) AS sum_v FROM (
							SELECT key AS k, value AS v FROM jsonb_each_text(coalesce(devices_domains_5s.domain, '{}'::jsonb))
							UNION ALL
							SELECT key, value FROM jsonb_each_text(EXCLUDED.domain)
						) x
						GROUP BY k
					) y
				),
				requests = devices_domains_5s.requests + EXCLUDED.requests`, cols, strings.Join(vals, ","))
		if err := exec(q, args...); err != nil {
			return err
		}
	}

	// Batch DeviceCountry
	if len(countries) > 0 {
		cols := "bucket,device_id,companies,countries,requests"
		var vals []string
		var args []interface{}
		for i, r := range countries {
			base := i * 5
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4, base+5))
			// store JSONB fields as strings and keep order: companies, countries
			args = append(args, r.Bucket, r.DeviceID, string(r.Company), string(r.Country), r.Requests)
		}
		q := fmt.Sprintf(`INSERT INTO devices_countries_5s (%s) VALUES %s
			ON CONFLICT (device_id, bucket) DO UPDATE
			SET companies = (
				SELECT jsonb_object_agg(k, to_jsonb(sum_v)) FROM (
					SELECT k, sum(v::bigint) AS sum_v FROM (
						SELECT key AS k, value AS v FROM jsonb_each_text(coalesce(devices_countries_5s.companies, '{}'::jsonb))
						UNION ALL
						SELECT key, value FROM jsonb_each_text(EXCLUDED.companies)
					) x
					GROUP BY k
				) y
			),
			countries = (
				SELECT jsonb_object_agg(k, to_jsonb(sum_v)) FROM (
					SELECT k, sum(v::bigint) AS sum_v FROM (
						SELECT key AS k, value AS v FROM jsonb_each_text(coalesce(devices_countries_5s.countries, '{}'::jsonb))
						UNION ALL
						SELECT key, value FROM jsonb_each_text(EXCLUDED.countries)
					) x
					GROUP BY k
				) y
			),
			requests = devices_countries_5s.requests + EXCLUDED.requests`, cols, strings.Join(vals, ","))
		if err := exec(q, args...); err != nil {
			return err
		}
	}

	// Batch DeviceProto
	if len(protos) > 0 {
		cols := "bucket,device_id,proto,requests"
		var vals []string
		var args []interface{}
		for i, r := range protos {
			base := i * 4
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4))
			args = append(args, r.Bucket, r.DeviceID, string(r.Proto), r.Requests)
		}
		q := fmt.Sprintf(`INSERT INTO devices_protos_5s (%s) VALUES %s
			ON CONFLICT (device_id, bucket) DO UPDATE
			SET proto = (
				SELECT jsonb_object_agg(k, to_jsonb(sum_v)) FROM (
					SELECT k, sum(v::bigint) AS sum_v FROM (
						SELECT key AS k, value AS v FROM jsonb_each_text(coalesce(devices_protos_5s.proto, '{}'::jsonb))
						UNION ALL
						SELECT key, value FROM jsonb_each_text(EXCLUDED.proto)
					) x
					GROUP BY k
				) y
			),
			requests = devices_protos_5s.requests + EXCLUDED.requests`, cols, strings.Join(vals, ","))
		if err := exec(q, args...); err != nil {
			return err
		}
	}

	// Update day cache versions: increment by 1 for each distinct day we modified.
	// Collect unique days from all record types (bucket -> date string YYYY-MM-DD).
	uniqueDays := make(map[string]struct{})
	for _, r := range traffics {
		uniqueDays[r.Bucket.UTC().Format("2006-01-02")] = struct{}{}
	}
	for _, r := range domains {
		uniqueDays[r.Bucket.UTC().Format("2006-01-02")] = struct{}{}
	}
	for _, r := range countries {
		uniqueDays[r.Bucket.UTC().Format("2006-01-02")] = struct{}{}
	}
	for _, r := range protos {
		uniqueDays[r.Bucket.UTC().Format("2006-01-02")] = struct{}{}
	}

	if len(uniqueDays) > 0 {
		var vals []string
		var args []interface{}
		i := 0
		for day := range uniqueDays {
			i++
			vals = append(vals, fmt.Sprintf("($%d,$%d)", (i-1)*2+1, (i-1)*2+2))
			args = append(args, day, 1)
		}
		q := fmt.Sprintf(`INSERT INTO day_cache_versions (day, version) VALUES %s
			ON CONFLICT (day) DO UPDATE SET version = day_cache_versions.version + 1`, strings.Join(vals, ","))
		if err := exec(q, args...); err != nil {
			return err
		}
	}

	return nil
}
