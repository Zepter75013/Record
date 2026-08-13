package stats

type Stats struct {
	Vinyls    int `json:"vinyls"`
	Artists   int `json:"artists"`
	Genres    int `json:"genres"`
	Formats   int `json:"formats"`
	Countries int `json:"countries"`
	Labels    int `json:"labels"`
}
