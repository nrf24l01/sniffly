package batcher

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

func (c *CHBatch) Insert(ctx context.Context, b *Batcher) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 4)

	// traffic worker
	if len(c.DeviceTraffics) > 0 {
		wg.Add(1)
		go func(rows []DeviceTraffic) {
			defer wg.Done()
			// Group by device_id and aggregate per bucket
			byDevice := make(map[uint64]map[time.Time]DeviceTraffic)
			for _, row := range rows {
				m, ok := byDevice[row.DeviceID]
				if !ok {
					m = make(map[time.Time]DeviceTraffic)
					byDevice[row.DeviceID] = m
				}
				if existing, ok := m[row.Bucket]; ok {
					existing.UpBytes += row.UpBytes
					existing.ReqCount += row.ReqCount
					m[row.Bucket] = existing
				} else {
					m[row.Bucket] = row
				}
			}
			for deviceID, bucketMap := range byDevice {
				times := make([]time.Time, 0, len(bucketMap))
				for t := range bucketMap {
					times = append(times, t)
				}
				have, notHave, err := b.checkAlreadyHadTimeBatches(ctx, times, "device_traffic_5s", deviceID)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				for _, t := range have {
					row := bucketMap[t]
					if err := b.CHDB.CH.Exec(ctx, "ALTER TABLE device_traffic_5s UPDATE up_bytes = up_bytes + ?, req_count = req_count + ? WHERE device_id = ? AND bucket = ?", row.UpBytes, row.ReqCount, deviceID, t); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
				}
				if len(notHave) > 0 {
					batchIns, err := b.CHDB.CH.PrepareBatch(ctx, "INSERT INTO device_traffic_5s (device_id, bucket, up_bytes, req_count) VALUES")
					if err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
					for _, t := range notHave {
						row := bucketMap[t]
						if err := batchIns.Append(row.DeviceID, row.Bucket, row.UpBytes, row.ReqCount); err != nil {
							select {
							case errCh <- err:
							default:
							}
							return
						}
					}
					if err := batchIns.Send(); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
				}
			}
		}(c.DeviceTraffics)
	}

	// domains worker
	if len(c.DeviceDomains) > 0 {
		wg.Add(1)
		go func(rows []DeviceDomain) {
			defer wg.Done()
			deviceID := rows[0].DeviceID
			m := make(map[time.Time]DeviceDomain)
			times := make([]time.Time, 0, len(rows))
			for _, r := range rows {
				m[r.Bucket] = r
				times = append(times, r.Bucket)
			}
			have, notHave, err := b.checkAlreadyHadTimeBatches(ctx, times, "device_domain_5s", deviceID)
			if err != nil {
				select { case errCh <- err: default: }
				return
			}
			for _, t := range have {
				newR := m[t]
				qr, err := b.CHDB.CH.Query(ctx, "SELECT domain, requests FROM device_domain_5s WHERE device_id = ? AND bucket = ?", deviceID, t)
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				var existingDomain string
				var existingReq uint64
				if qr.Next() {
					if err := qr.Scan(&existingDomain, &existingReq); err != nil {
						qr.Close()
						select { case errCh <- err: default: }
						return
					}
				}
				qr.Close()
				var existMap map[string]uint64
				var newMap map[string]uint64
				if existingDomain != "" {
					_ = json.Unmarshal([]byte(existingDomain), &existMap)
				}
				_ = json.Unmarshal([]byte(newR.Domain), &newMap)
				if existMap == nil {
					existMap = make(map[string]uint64)
				}
				for k, v := range newMap {
					existMap[k] += v
				}
				merged, err := json.Marshal(existMap)
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				if err := b.CHDB.CH.Exec(ctx, "ALTER TABLE device_domain_5s UPDATE domain = ?, requests = requests + ? WHERE device_id = ? AND bucket = ?", string(merged), newR.Requests, deviceID, t); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
			if len(notHave) > 0 {
				batchIns, err := b.CHDB.CH.PrepareBatch(ctx, "INSERT INTO device_domain_5s (device_id, bucket, domain, requests) VALUES")
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				for _, t := range notHave {
					r := m[t]
					if err := batchIns.Append(r.DeviceID, r.Bucket, r.Domain, r.Requests); err != nil {
						select { case errCh <- err: default: }
						return
					}
				}
				if err := batchIns.Send(); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
		}(c.DeviceDomains)
	}

	// countries worker
	if len(c.DeviceCountries) > 0 {
		wg.Add(1)
		go func(rows []DeviceCountry) {
			defer wg.Done()
			deviceID := rows[0].DeviceID
			m := make(map[time.Time]DeviceCountry)
			times := make([]time.Time, 0, len(rows))
			for _, r := range rows {
				m[r.Bucket] = r
				times = append(times, r.Bucket)
			}
			have, notHave, err := b.checkAlreadyHadTimeBatches(ctx, times, "device_country_5s", deviceID)
			if err != nil {
				select { case errCh <- err: default: }
				return
			}
			for _, t := range have {
				newR := m[t]
				qr, err := b.CHDB.CH.Query(ctx, "SELECT companies, countries, requests FROM device_country_5s WHERE device_id = ? AND bucket = ?", deviceID, t)
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				var existCompanies []string
				var existCountries []string
				var existReq uint64
				if qr.Next() {
					if err := qr.Scan(&existCompanies, &existCountries, &existReq); err != nil {
						qr.Close()
						select { case errCh <- err: default: }
						return
					}
				}
				qr.Close()
				comps := make(map[string]struct{})
				for _, c := range existCompanies {
					comps[c] = struct{}{}
				}
				for _, c := range newR.Company {
					comps[c] = struct{}{}
				}
				mergedComps := make([]string, 0, len(comps))
				for k := range comps {
					mergedComps = append(mergedComps, k)
				}
				cnts := make(map[string]struct{})
				for _, c := range existCountries {
					cnts[c] = struct{}{}
				}
				for _, c := range newR.Country {
					cnts[c] = struct{}{}
				}
				mergedCnts := make([]string, 0, len(cnts))
				for k := range cnts {
					mergedCnts = append(mergedCnts, k)
				}
				if err := b.CHDB.CH.Exec(ctx, "ALTER TABLE device_country_5s UPDATE companies = ?, countries = ?, requests = requests + ? WHERE device_id = ? AND bucket = ?", mergedComps, mergedCnts, newR.Requests, deviceID, t); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
			if len(notHave) > 0 {
				batchIns, err := b.CHDB.CH.PrepareBatch(ctx, "INSERT INTO device_country_5s (device_id, bucket, companies, countries, requests) VALUES")
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				for _, t := range notHave {
					r := m[t]
					if err := batchIns.Append(r.DeviceID, r.Bucket, r.Company, r.Country, r.Requests); err != nil {
						select { case errCh <- err: default: }
						return
					}
				}
				if err := batchIns.Send(); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
		}(c.DeviceCountries)
	}

	// protos worker
	if len(c.DeviceProtos) > 0 {
		wg.Add(1)
		go func(rows []DeviceProto) {
			defer wg.Done()
			deviceID := rows[0].DeviceID
			m := make(map[time.Time]DeviceProto)
			times := make([]time.Time, 0, len(rows))
			for _, r := range rows {
				m[r.Bucket] = r
				times = append(times, r.Bucket)
			}
			have, notHave, err := b.checkAlreadyHadTimeBatches(ctx, times, "device_proto_5s", deviceID)
			if err != nil {
				select { case errCh <- err: default: }
				return
			}
			for _, t := range have {
				newR := m[t]
				qr, err := b.CHDB.CH.Query(ctx, "SELECT proto, requests FROM device_proto_5s WHERE device_id = ? AND bucket = ?", deviceID, t)
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				var existingProto string
				var existingReq uint64
				if qr.Next() {
					if err := qr.Scan(&existingProto, &existingReq); err != nil {
						qr.Close()
						select { case errCh <- err: default: }
						return
					}
				}
				qr.Close()
				var existMap map[string]uint64
				var newMap map[string]uint64
				if existingProto != "" {
					_ = json.Unmarshal([]byte(existingProto), &existMap)
				}
				_ = json.Unmarshal([]byte(newR.Proto), &newMap)
				if existMap == nil {
					existMap = make(map[string]uint64)
				}
				for k, v := range newMap {
					existMap[k] += v
				}
				merged, err := json.Marshal(existMap)
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				if err := b.CHDB.CH.Exec(ctx, "ALTER TABLE device_proto_5s UPDATE proto = ?, requests = requests + ? WHERE device_id = ? AND bucket = ?", string(merged), newR.Requests, deviceID, t); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
			if len(notHave) > 0 {
				batchIns, err := b.CHDB.CH.PrepareBatch(ctx, "INSERT INTO device_proto_5s (device_id, bucket, proto, requests) VALUES")
				if err != nil {
					select { case errCh <- err: default: }
					return
				}
				for _, t := range notHave {
					r := m[t]
					if err := batchIns.Append(r.DeviceID, r.Bucket, r.Proto, r.Requests); err != nil {
						select { case errCh <- err: default: }
						return
					}
				}
				if err := batchIns.Send(); err != nil {
					select { case errCh <- err: default: }
					return
				}
			}
		}(c.DeviceProtos)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Batcher) checkAlreadyHadTimeBatches(ctx context.Context, times []time.Time, table_name string, device_id uint64) ([]time.Time, []time.Time, error) {
	if len(times) == 0 {
		return nil, nil, nil
	}
	// Use a persistent helper table to avoid temporary-table/session API requirements.
	// helper_buckets(session_id UInt64, bucket DateTime)

	// Ensure helper table exists (idempotent)
	if err := b.CHDB.CH.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS helper_buckets (
			session_id UInt64,
			bucket DateTime
		) ENGINE = MergeTree ORDER BY (session_id, bucket)
	`); err != nil {
		return nil, nil, err
	}

	// Create a session id for this operation
	var sessionID uint64
	if b.SnowflakeNode != nil {
		sessionID = uint64(b.SnowflakeNode.Generate().Int64())
	} else {
		sessionID = uint64(time.Now().UnixNano())
	}

	// Insert requested buckets for this session into helper_buckets
	if len(times) > 0 {
		batchIns, err := b.CHDB.CH.PrepareBatch(ctx, "INSERT INTO helper_buckets (session_id, bucket) VALUES")
		if err != nil {
			return nil, nil, err
		}
		for _, t := range times {
			if err := batchIns.Append(sessionID, t); err != nil {
				return nil, nil, err
			}
		}
		if err := batchIns.Send(); err != nil {
			return nil, nil, err
		}
	}

	// Join the helper table with the target table to find which buckets already exist for this device
	query := "SELECT DISTINCT d.bucket FROM " + table_name + " AS d INNER JOIN helper_buckets AS h ON d.bucket = h.bucket WHERE d.device_id = ? AND h.session_id = ?"
	rows, err := b.CHDB.CH.Query(ctx, query, device_id, sessionID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	present := make(map[time.Time]struct{})
	for rows.Next() {
		var bucket time.Time
		if err := rows.Scan(&bucket); err != nil {
			return nil, nil, err
		}
		present[bucket] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	have := make([]time.Time, 0, len(times))
	notHave := make([]time.Time, 0, len(times))
	for _, t := range times {
		if _, ok := present[t]; ok {
			have = append(have, t)
		} else {
			notHave = append(notHave, t)
		}
	}

	// Cleanup helper rows for this session. Use ALTER TABLE ... DELETE WHERE to remove session rows.
	// Note: ClickHouse DELETE is asynchronous depending on the engine and settings.
	if err := b.CHDB.CH.Exec(ctx, "ALTER TABLE helper_buckets DELETE WHERE session_id = ?", sessionID); err != nil {
		// non-fatal: return results even if cleanup fails, but surface the error
		return have, notHave, err
	}

	return have, notHave, nil
}