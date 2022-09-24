package actuator

func ExampleSetHealth() {
	SetHealth("Printer", Down, "no more ink")
}
