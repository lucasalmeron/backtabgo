package card

type Card struct {
	ID             int      `json:"id"`
	Word           string   `json:"word"`
	ForbbidenWords []string `json:"forbbidenWords"`
}
