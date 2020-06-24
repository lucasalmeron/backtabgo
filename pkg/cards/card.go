package card

type Card struct {
	ID             string   `json:"id" bson:"_id,omitempty"`
	Word           string   `json:"word" bson:"word"`
	ForbbidenWords []string `json:"forbbidenWords" bson:"forbbidenWords"`
}
