package service

import "time"

var BackoffSchedule = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
	15 * time.Second,
	60 * time.Second,
}
