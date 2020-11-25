package common

import (
	"github.com/woodliu/PrometheusMiddleware/config"
)

/*
const (
	apiPrefix = "/api/v1"

	epAlerts          = apiPrefix + "/alerts"
	epAlertManagers   = apiPrefix + "/alertmanagers"
	epQuery           = apiPrefix + "/query"
	epQueryRange      = apiPrefix + "/query_range"
	epLabels          = apiPrefix + "/labels"
	epLabelValues     = apiPrefix + "/label/:name/values"
	epSeries          = apiPrefix + "/series"
	epTargets         = apiPrefix + "/targets"
	epTargetsMetadata = apiPrefix + "/targets/metadata"
	epMetadata        = apiPrefix + "/metadata"
	epRules           = apiPrefix + "/rules"
	epSnapshot        = apiPrefix + "/admin/tsdb/snapshot"
	epDeleteSeries    = apiPrefix + "/admin/tsdb/delete_series"
	epCleanTombstones = apiPrefix + "/admin/tsdb/clean_tombstones"
	epConfig          = apiPrefix + "/status/config"
	epFlags           = apiPrefix + "/status/flags"
	epRuntimeinfo     = apiPrefix + "/status/runtimeinfo"
	epTSDB            = apiPrefix + "/status/tsdb"
)
*/

const(
	Query = iota + 1
	QueryRange
	Labels
	LabelName
	Series
)

/* errors */
const (
	InvalidMetricErr = "invalid metric"
	InvalidTimestampErr = "invalid timestamp"
	TooEarlyTimestampErr = "too early timestamp"
	InvalidTimeRangeErr = "invalid time range"
	InvalidSampleNumErr = "invalid sample number"
	InvalidPrecisionErr = "invalid precision"
	ForbiddenMetricErr = "forbidden metric" /* for query and query_range*/
	QueryPrometheusErr = "query prometheus error"
	InvalidQueryTypeErr = "invalid query type"
	NoRawQueryErr = "request has no raw query"
	InvalidParametersStructErr = "invalid parameters struct"
	TooManyRequestsErr = "too many requests"
)

func InitEnv() error{
	return config.LoadConfig(config.ConfigFilePath)
}