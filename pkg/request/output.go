package request

type (
	Ress struct{
		Result []Res `json:"result"`
	}

	Res struct {
		Value string		`json:"value"`
		Timestamp string	`json:"timestamp"`
	}
)
