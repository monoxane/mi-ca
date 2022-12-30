package model

type Cluster struct {
	Name          string   `json:"name"`
	Project       string   `json:"project"`
	Workers       int      `json:"workers"`
	ActiveWorkers int      `json:"active_workers"`
	Concurrency   int      `json:"concurrency"`
	Owners        []string `json:"owners"`
	Cause         string   `json:"cause"`
}
