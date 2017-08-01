package importer

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/pingcap/tidb/mysql"
)

var defaultStep int64 = 1

type datum struct {
	sync.Mutex

	intValue    int64
	minIntValue int64
	maxIntValue int64
	timeValue   time.Time
	step        int64

	init     bool
	useRange bool
}

func newDatum() *datum {
	return &datum{intValue: -1, step: 1}
}

func (d *datum) uniqString(n int) string {
	d.Lock()
	d.intValue++
	data := d.intValue
	d.Unlock()

	var value []byte
	for ; ; n-- {
		if n == 0 {
			break
		}

		idx := data % int64(len(alphabet))
		data = data / int64(len(alphabet))

		value = append(value, alphabet[idx])

		if data == 0 {
			break
		}
	}

	for i, j := 0, len(value)-1; i < j; i, j = i+1, j-1 {
		value[i], value[j] = value[j], value[i]
	}

	return string(value)
}

func (d *datum) uniqDate() string {
	d.Lock()
	defer d.Unlock()

	if d.timeValue.IsZero() {
		d.timeValue = time.Now()
	} else {
		d.timeValue = d.timeValue.AddDate(0, 0, int(d.step))
	}

	return fmt.Sprintf("%04d-%02d-%02d", d.timeValue.Year(), d.timeValue.Month(), d.timeValue.Day())
}

func (d *datum) uniqTimestamp() string {
	d.Lock()
	defer d.Unlock()

	if d.timeValue.IsZero() {
		d.timeValue = time.Now()
	} else {
		d.timeValue = d.timeValue.Add(time.Duration(d.step) * time.Second)
	}

	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		d.timeValue.Year(), d.timeValue.Month(), d.timeValue.Day(),
		d.timeValue.Hour(), d.timeValue.Minute(), d.timeValue.Second())
}

func (d *datum) uniqTime() string {
	d.Lock()
	defer d.Unlock()

	if d.timeValue.IsZero() {
		d.timeValue = time.Now()
	} else {
		d.timeValue = d.timeValue.Add(time.Duration(d.step) * time.Second)
	}

	return fmt.Sprintf("%02d:%02d:%02d", d.timeValue.Hour(), d.timeValue.Minute(), d.timeValue.Second())
}

func (d *datum) uniqYear() string {
	d.Lock()
	defer d.Unlock()

	if d.timeValue.IsZero() {
		d.timeValue = time.Now()
	} else {
		d.timeValue = d.timeValue.AddDate(int(d.step), 0, 0)
	}

	return fmt.Sprintf("%04d", d.timeValue.Year())
}

// Initialize int datum
func (d *datum) setInitInt64Value(step int64, min int64, max int64) {
	d.Lock()
	defer d.Unlock()

	if d.init {
		return
	}

	d.step = step

	if min != -1 {
		d.minIntValue = min
		d.intValue = min
	}

	if min < max {
		d.maxIntValue = max
		d.useRange = true
	}

	d.init = true
}

func (d *datum) uniqInt64() int64 {
	d.Lock()
	defer d.Unlock()

	data := d.intValue
	if d.useRange {
		if d.intValue+d.step > d.maxIntValue {
			return data
		}
	}

	d.intValue += d.step
	return data
}

// Generate row data
func GenRowData(table *Table) (string, error) {
	var values []byte
	for _, column := range table.columns {
		data, err := genColumnData(table, column)
		if err != nil {
			return "", errors.Trace(err)
		}
		values = append(values, []byte(data)...)
		values = append(values, ',')
	}

	// omits last comma
	values = values[:len(values)-1]
	sql := fmt.Sprintf("insert into %s (%s) values (%s);", table.name, table.columnList, string(values))
	return sql, nil
}

func genColumnData(table *Table, column *column) (string, error) {
	tp := column.tp
	_, isUnique := table.uniqIndices[column.name]
	isUnsigned := mysql.HasUnsignedFlag(tp.Flag)

	switch tp.Tp {
	case mysql.TypeTiny:
		var data int64
		if isUnique {
			data = uniqInt64Value(column, 0, math.MaxUint8)
		} else {
			if isUnsigned {
				data = randInt64Value(column, 0, math.MaxUint8)
			} else {
				data = randInt64Value(column, math.MinInt8, math.MaxInt8)
			}
		}
		return strconv.FormatInt(data, 10), nil
	case mysql.TypeShort:
		var data int64
		if isUnique {
			data = uniqInt64Value(column, 0, math.MaxUint16)
		} else {
			if isUnsigned {
				data = randInt64Value(column, 0, math.MaxUint16)
			} else {
				data = randInt64Value(column, math.MinInt16, math.MaxInt16)
			}
		}
		return strconv.FormatInt(data, 10), nil
	case mysql.TypeLong:
		var data int64
		if isUnique {
			data = uniqInt64Value(column, 0, math.MaxUint32)
		} else {
			if isUnsigned {
				data = randInt64Value(column, 0, math.MaxUint32)
			} else {
				data = randInt64Value(column, math.MinInt32, math.MaxInt32)
			}
		}
		return strconv.FormatInt(data, 10), nil
	case mysql.TypeLonglong:
		var data int64
		if isUnique {
			data = uniqInt64Value(column, 0, math.MaxInt64)
		} else {
			if isUnsigned {
				data = randInt64Value(column, 0, math.MaxInt64)
			} else {
				data = randInt64Value(column, math.MinInt32, math.MaxInt32)
			}
		}
		return strconv.FormatInt(data, 64), nil
	case mysql.TypeVarchar, mysql.TypeString, mysql.TypeTinyBlob, mysql.TypeBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
		data := []byte{'\''}
		if isUnique {
			data = append(data, []byte(column.data.uniqString(tp.Flen))...)
		} else {
			data = append(data, []byte(randString(randInt(1, tp.Flen)))...)
		}

		data = append(data, '\'')
		return string(data), nil
	case mysql.TypeFloat, mysql.TypeDouble, mysql.TypeDecimal:
		var data float64
		if isUnique {
			data = float64(uniqInt64Value(column, 0, math.MaxInt64))
		} else {
			if isUnsigned {
				data = float64(randInt64Value(column, 0, math.MaxInt64))
			} else {
				data = float64(randInt64Value(column, math.MinInt32, math.MaxInt32))
			}
		}
		return strconv.FormatFloat(data, 'f', -1, 64), nil
	case mysql.TypeDate:
		data := []byte{'\''}
		if isUnique {
			data = append(data, []byte(column.data.uniqDate())...)
		} else {
			data = append(data, []byte(randDate(column.min, column.max))...)
		}

		data = append(data, '\'')
		return string(data), nil
	case mysql.TypeDatetime, mysql.TypeTimestamp:
		data := []byte{'\''}
		if isUnique {
			data = append(data, []byte(column.data.uniqTimestamp())...)
		} else {
			data = append(data, []byte(randTimestamp(column.min, column.max))...)
		}

		data = append(data, '\'')
		return string(data), nil
	case mysql.TypeDuration:
		data := []byte{'\''}
		if isUnique {
			data = append(data, []byte(column.data.uniqTime())...)
		} else {
			data = append(data, []byte(randTime(column.min, column.max))...)
		}

		data = append(data, '\'')
		return string(data), nil
	case mysql.TypeYear:
		data := []byte{'\''}
		if isUnique {
			data = append(data, []byte(column.data.uniqYear())...)
		} else {
			data = append(data, []byte(randYear(column.min, column.max))...)
		}

		data = append(data, '\'')
		return string(data), nil
	default:
		return "", errors.Errorf("unsupported column type - %v", column)
	}

}

func randInt64Value(column *column, min int64, max int64) int64 {
	if len(column.set) > 0 {
		idx := randInt(0, len(column.set)-1)
		data, _ := strconv.ParseInt(column.set[idx], 10, 64)
		return data
	}

	min, max = intRangeValue(column, min, max)
	return randInt64(min, max)
}

func intRangeValue(column *column, min int64, max int64) (int64, int64) {
	var err error
	if len(column.min) > 0 {
		min, err = strconv.ParseInt(column.min, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		if len(column.max) > 0 {
			max, err = strconv.ParseInt(column.max, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return min, max
}

func uniqInt64Value(column *column, min int64, max int64) int64 {
	min, max = intRangeValue(column, min, max)
	// initialize datum, notice race condition here
	column.data.setInitInt64Value(column.step, min, max)
	return column.data.uniqInt64()
}

// reference: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func randString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMax); idx < len(alphabet) {
			b[i] = alphabet[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
