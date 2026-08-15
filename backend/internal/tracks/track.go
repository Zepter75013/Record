package tracks

type Track struct {
	ID         int    `json:"id"`
	VinylID    int    `json:"vinyl_id"`
	TrackOrder int    `json:"track_order"`
	Position   string `json:"position"`
	Title      string `json:"title"`
	Duration   string `json:"duration"`
}
