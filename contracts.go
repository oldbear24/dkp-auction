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

type ChangeTokens struct {
	User   string `json:"user"`
	Amount int    `json:"amount"`
}

type Settings struct {
	NameSynchronization           bool   `db:"nameSynchronization"`
	SynchronizationType           string `db:"synchronizationType"`
	SynchronizationUrl            string `db:"synchronizationUrl"`
	SynchronizationClient         string `db:"synchronizationClient"`
	SynchronizationPassword       string `db:"synchronizationPassword"`
	SynchronizationDiscordGuildId string `db:"synchronizationDiscordGuildId"`
	EnableFloatingEndOfAuction    bool   `db:"enableFloatingEndOfAuction"`
	FloatingEndOfAuctionMinutes   int    `db:"floatingEndOfAuctionMinutes"`
}
