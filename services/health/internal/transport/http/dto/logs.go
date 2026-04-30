package dto

type MetricValueRequest struct {
	MetricID string  `json:"metric_id"`
	ValueNum float64 `json:"value_num"`
}

type CreateOrUpdateLogRequest struct {
	OccurredAt   string               `json:"occurred_at"`
	LogTypeID    *string              `json:"log_type_id"`
	Description  *string              `json:"description"`
	MetricValues []MetricValueRequest `json:"metric_values"`
	Attachments  []AttachmentRequest  `json:"attachments"`
	RowVersion   int                  `json:"row_version"`
}

type DeleteLogRequest struct {
	RowVersion int `json:"row_version"`
}

type LogTypeMetricRequirementRequest struct {
	MetricID   string `json:"metric_id"`
	IsRequired bool   `json:"is_required"`
}

type CreateLogTypeRequest struct {
	Name               string                            `json:"name"`
	MetricRequirements []LogTypeMetricRequirementRequest `json:"metric_requirements"`
}

type UpdateLogTypeRequest struct {
	Name               string                            `json:"name"`
	MetricRequirements []LogTypeMetricRequirementRequest `json:"metric_requirements"`
	RowVersion         int                               `json:"row_version"`
}

type CreateMetricRequest struct {
	Name      string   `json:"name"`
	InputKind string   `json:"input_kind"`
	Unit      *string  `json:"unit"`
	MinValue  *float64 `json:"min_value"`
	MaxValue  *float64 `json:"max_value"`
}

type UpdateMetricRequest struct {
	Name       string   `json:"name"`
	InputKind  string   `json:"input_kind"`
	Unit       *string  `json:"unit"`
	MinValue   *float64 `json:"min_value"`
	MaxValue   *float64 `json:"max_value"`
	RowVersion int      `json:"row_version"`
}
