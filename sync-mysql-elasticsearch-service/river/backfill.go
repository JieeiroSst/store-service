package river

import (
	"fmt"
	"strings"
)

func (r *River) Backfill() error {
	for _, rule := range r.rules {
		if err := r.backfillRule(rule); err != nil {
			return fmt.Errorf("backfill %s.%s: %w", rule.Schema, rule.Table, err)
		}
	}
	return nil
}

func (r *River) backfillRule(rule *Rule) error {
	columns := make([]string, 0, len(rule.TableInfo.Columns))
	for _, c := range rule.TableInfo.Columns {
		columns = append(columns, "`"+c.Name+"`")
	}

	query := fmt.Sprintf("SELECT %s FROM `%s`.`%s`", strings.Join(columns, ","), rule.Schema, rule.Table)

	res, err := r.canal.Execute(query)
	if err != nil {
		return err
	}
	defer res.Close()

	if res.Resultset == nil {
		return nil
	}

	bulkSize := r.c.BulkSize
	if bulkSize <= 0 {
		bulkSize = 128
	}

	batch := make([][]interface{}, 0, bulkSize)
	for rowIdx := 0; rowIdx < res.Resultset.RowNumber(); rowIdx++ {
		row := make([]interface{}, len(rule.TableInfo.Columns))
		for colIdx := range rule.TableInfo.Columns {
			v, err := res.GetValue(rowIdx, colIdx)
			if err != nil {
				return err
			}
			// makeReqColumnData expects string (not []byte) for
			// text/date/enum/set/bit columns, matching the mysqldump path.
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			row[colIdx] = v
		}
		batch = append(batch, row)

		if len(batch) >= bulkSize {
			if err := r.flushBackfillBatch(rule, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		return r.flushBackfillBatch(rule, batch)
	}
	return nil
}

func (r *River) flushBackfillBatch(rule *Rule, rows [][]interface{}) error {
	reqs, err := r.makeInsertRequest(rule, rows)
	if err != nil {
		return err
	}
	return r.doBulk(reqs)
}
