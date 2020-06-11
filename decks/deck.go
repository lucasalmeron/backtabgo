package deck

import (
	card "github.com/lucasalmeron/backtabgo"
)

type Deck struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Theme   string    `json:"theme"`
	Cards card[] `json:"cards"`
}