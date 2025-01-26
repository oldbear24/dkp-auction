package main

type BidStruct struct {
	Amount int `json:"amount"`
}

type TokenHealtCheck struct {
	State       string                `json:"state"`
	UserResults []TokenHealtCheckUser `json:"userResults"`
}
type TokenHealtCheckUser struct {
	State             string `json:"state"`
	User              string `json:"user"`
	UserTokens        int    `json:"userTokens"`
	TransactionTokens int    `json:"transactionTokens"`
	Differece         int    `json:"differece"`
}
