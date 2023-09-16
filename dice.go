package main

import (
	"dice/html"
	"errors"
	"math/rand"
)

var activeTables = make(map[string]*table, 10)

type player struct {
	Name      string
	Index     int
	BetAmount int
	Bet       string
	Wallet    int
	LastRoll  []uint
	IsShooter bool
}

type table struct {
	Name          string
	InternalName  string
	Players       []player
	Point         uint
	Complete      bool
	BetHight      uint
	MaxBetHight   uint
	wsConnections map[*Client]bool
}

func (p *player) roll(determineShooter bool) {
	if determineShooter {
		p.placeBet(10)
	}
	p.LastRoll[0] = uint(rand.Intn(6-1) + 1)
	p.LastRoll[1] = uint(rand.Intn(6-1) + 1)
}

func (p *player) setWager(bet string) {
	p.Bet = bet
}

func (p *player) placeBet(amount int) {
	if amount == 0 {
		amount = 10
	}

	if p.Wallet-amount < 0 {
		return
	}
	p.BetAmount = amount
	p.Wallet -= amount
}

func (t *table) broadcastGameState() {
	type gameState struct {
		Table  table
		Player player
	}

	for c := range t.wsConnections {
		c.send <- html.WSGameState(gameState{
			Table:  *t,
			Player: *c.player,
		})
	}
}

func (t *table) determineShooter() {
	var (
		highestRoll int
		sum         uint
		maxBetHight int
	)

	for i, p := range t.Players {
		if p.LastRoll[0] == 0 && p.LastRoll[1] == 0 {
			return
		}

		if (p.LastRoll[0] + p.LastRoll[1]) > sum {
			sum = p.LastRoll[0] + p.LastRoll[1]
			highestRoll = i
		}

		if p.Wallet < maxBetHight || maxBetHight == 0 {
			maxBetHight = p.Wallet
		}
	}

	t.Players[highestRoll].IsShooter = true
	t.MaxBetHight = uint(maxBetHight)

	for c := range t.wsConnections {
		if c.player.IsShooter {
			c.send <- html.ShowWagerControlls(10)
		} else {
			c.send <- []byte(`<div id="player-control">The shooter is setting te bet </div>`)
		}
	}

	t.broadcastGameState()
}

func (t *table) evaluateRoll() {
	var result string
	var rolled uint
	for _, p := range t.Players {
		if p.IsShooter {
			rolled = p.LastRoll[0] + p.LastRoll[1]
		}
	}

	if t.Point == 0 {
		result = passOrCraps(rolled)
		if result == "" {
			t.Point = rolled
		}
	}
}

func (t *table) payout(result string) {
}

func (t *table) setForNextRound() {
	t.Point = 0
	t.BetHight = 0
	t.MaxBetHight = 0
	for _, p := range t.Players {
		p.LastRoll[0] = 0
		p.LastRoll[1] = 0
		p.IsShooter = false
		p.BetAmount = 0
	}
}

func (t *table) letNonShootersBet() {
	for c := range t.wsConnections {
		if !c.player.IsShooter {
			c.send <- html.ShowWagerControlls(t.BetHight)
		}
	}
}

func newTable(tableName, playerName string) (string, error) {
	if len(activeTables) >= 10 {
		return "", errors.New("Max tables reached.")
	}

	t := table{}
	p := newPlayer(playerName)
	t.Name = tableName
	t.InternalName = createRandomString()
	t.Players = append(t.Players, p)
	t.wsConnections = make(map[*Client]bool)

	activeTables[t.InternalName] = &t

	return t.InternalName, nil
}

func newPlayer(playerName string) player {
	return player{
		Name:     playerName,
		Wallet:   110,
		LastRoll: make([]uint, 2),
	}
}

func passOrCraps(rolled uint) string {
	if rolled == 7 || rolled == 11 {
		return "pass"
	}
	if rolled == 2 || rolled == 3 || rolled == 12 {
		return "craps"
	}

	return ""
}
