package propagation

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestCarrier(t *testing.T) {
	x := amqp.Table(make(map[string]interface{}))
	x["a"] = amqp.Table(make(map[string]interface{}))
	x["d"] = "foo"

	HeaderCarrier(x).Set("a.b", "c")
	if HeaderCarrier(x).Get("a.b") != "c" {
		t.Fail()
	}
	if x["a"].(amqp.Table)["b"] != "c" {
		t.Fail()
	}

	keys := HeaderCarrier(x).Keys()
	if len(keys) > 2 {
		t.Fail()
	}
	for _, key := range keys {
		if key != "a.b" && key != "d" {
			t.Fail()
		}
	}
}
