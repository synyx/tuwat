package propagation

import (
	"strconv"
	"strings"

	"github.com/streadway/amqp"
)

// HeaderCarrier adapts amqp.Table to satisfy the TextMapCarrier interface.
type HeaderCarrier amqp.Table

// Get returns the value associated with the passed key.
func (hc HeaderCarrier) Get(key string) string {
	path := strings.Split(key, ".")
	tbl := amqp.Table(hc)
	for ; len(path) > 1; path = path[1:] {
		if t, ok := tbl[path[0]].(amqp.Table); ok {
			tbl = t
		}
	}

	if v, ok := tbl[path[0]]; ok {
		switch value := v.(type) {
		case string:
			return value
		case bool:
			if value {
				return "1"
			} else {
				return "0"
			}
		case int:
		case int8:
		case int16:
		case int64:
		case int32:
			return strconv.FormatInt(int64(value), 10)
		case uint:
		case uint8:
		case uint16:
		case uint64:
		case uint32:
			return strconv.FormatUint(uint64(value), 10)
		case float32:
		case float64:
			return strconv.FormatFloat(float64(value), 'g', -1, 64)
		}
	}
	return ""
}

// Set stores the key-value pair.
func (hc HeaderCarrier) Set(key string, value string) {
	path := strings.Split(key, ".")
	tbl := amqp.Table(hc)
	for ; len(path) > 1; path = path[1:] {
		if t, ok := tbl[path[0]].(amqp.Table); ok {
			tbl = t
		}
	}
	tbl[path[0]] = value
}

// Keys lists the keys stored in this carrier.
func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k, v := range hc {
		switch value := v.(type) {
		case string:
			keys = append(keys, k)
		case amqp.Table:
			for _, k2 := range HeaderCarrier(value).Keys() {
				keys = append(keys, strings.Join([]string{k, k2}, "."))
			}
		case bool:
			keys = append(keys, k)
		case int:
		case int8:
		case int16:
		case int64:
		case int32:
		case uint:
		case uint8:
		case uint16:
		case uint64:
		case uint32:
		case float32:
		case float64:
			keys = append(keys, k)
		}
	}
	return keys
}
