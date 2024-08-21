package graylog

type alertResult struct {
	Total  int     `json:"total"`
	Alerts []alert `json:"alerts"`
}

type alert struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	ConditionId string `json:"condition_id"`
	StreamId    string `json:"stream_id"`
	TriggeredAt string `json:"triggered_at"`
	ResolvedAt  string `json:"resolved_at"`
}
