package repository

type StatsRepository interface {
	GetCounts() (map[string]int, error)
}
