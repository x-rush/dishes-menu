package dao

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON is a []byte-backed column type that auto-marshals/unmarshals JSON.
//
// It implements sql.Scanner and driver.Valuer, so sqlx can use it directly.
type JSON []byte

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = []byte(v)
		return nil
	default:
		return fmt.Errorf("dao.JSON: unsupported scan type %T", value)
	}
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("dao.JSON: invalid json: %s", string(data))
	}
	*j = append((*j)[:0], data...)
	return nil
}
