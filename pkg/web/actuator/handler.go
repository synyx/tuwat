package actuator

var HealthAggregator *HealthActuator

func init() {
	HealthAggregator = newHealthActuator()
}

func SetHealth(check string, status Status, message string) {
	HealthAggregator.Set(check, status, message)
}
