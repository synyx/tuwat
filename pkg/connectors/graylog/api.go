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

type eventsSearchParameters struct {
	Query     string             `json:"query"`
	Page      int                `json:"page"`
	PerPage   int                `json:"per_page"`
	Filter    eventsSearchFilter `json:"filter"`
	TimeRange timeRange          `json:"timerange"`
}

type eventsSearchFilter struct {
	Alerts AlertsFilerType `json:"alerts"`
}

type AlertsFilerType = string

const (
	AlertsFilterInclude AlertsFilerType = "include"
	AlertsFilterExclude AlertsFilerType = "exclude"
	AlertsFilterOnly    AlertsFilerType = "only"
)

type timeRange struct {
	Type  string `json:"type"`
	Range int    `json:"range"`
}

type timeRangeType = string

const (
	TimeRangeRelative timeRangeType = "relative"
)

type eventsSearchResults struct {
	Context eventsSearchResultContext
	Events  []eventsSearchResult
}

type eventsSearchResultContext struct {
	EventDefinitions map[string]eventDefinition  `json:"event_definitions"`
	Streams          map[string]streamDefinition `json:"streams"`
}

type eventDefinition struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	Title       string `json:"title"`
}

type streamDefinition struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
type eventsSearchResult struct {
	Event eventsSearchEventResult `json:"event"`
}
type eventsSearchEventResult struct {
	Id                  string              `json:"id"`
	EventDefinitionId   string              `json:"event_definition_id"`
	EventDefinitionType eventDefinitionType `json:"event_definition_type"`
	Alert               bool                `json:"alert"`
	Message             string              `json:"message"`
	Source              string              `json:"source"`
	TimeStamp           string              `json:"timestamp"`
	TimeRangeStart      string              `json:"timerange_start"`
	TimeRangeEnd        string              `json:"timerange_end"`
	Streams             []string            `json:"streams"`
	Priority            int                 `json:"priority"`
}

type eventDefinitionType = string

const (
	eventDefinitionAggregationv1 eventDefinitionType = "aggregation-v1"
	eventDefinitionCorrelationv1 eventDefinitionType = "correlation-v1"
)
