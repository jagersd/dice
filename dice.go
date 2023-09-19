package main

import (
	"dice/html"
	"errors"
	"math/rand"
)

var activeTables = make(map[string]*table, 10)

const minimumBet int = 5

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
		p.placeBet(minimumBet)
	}
	p.LastRoll[0] = uint(rand.Intn(6-1) + 1)
	p.LastRoll[1] = uint(rand.Intn(6-1) + 1)
}

func (p *player) setWager(bet string) {
	p.Bet = bet
}

func (p *player) placeBet(amount int) {
	if amount == 0 {
		amount = int(minimumBet)
	}

	if (p.Wallet - amount) < 1 {
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
		highestRoller int
		sum           uint
		maxBetHight   int
	)

	for i, p := range t.Players {
		if p.LastRoll[0] == 0 && p.LastRoll[1] == 0 {
			return
		}

		if (p.LastRoll[0] + p.LastRoll[1]) > sum {
			sum = p.LastRoll[0] + p.LastRoll[1]
			highestRoller = i
		}

		if p.Wallet < maxBetHight || maxBetHight == 0 {
			maxBetHight = p.Wallet
		}
	}

	t.Players[highestRoller].IsShooter = true
	t.MaxBetHight = uint(maxBetHight)

	for c := range t.wsConnections {
		if c.player.IsShooter {
			c.send <- html.ShowWagerControlls(minimumBet)
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
			return
		} else {
			t.payout(result)
			return
		}
	} else {
		if rolled == t.Point {
			t.payout("pass")
			return
		}
		if rolled == 7 {
			t.payout("craps")
			return
		}
	}
}

func (t *table) payout(result string) {
	var basePot, sidePot int
	var winners []int

	basePot = len(t.Players) * 5

	for c := range t.wsConnections {
		c.send <- []byte(`<div id="announcements"><h2>` + result + ` won! </h2></div>`)
		c.send <- html.Reset()
	}

	for i, p := range t.Players {
		basePot += p.BetAmount
		t.Players[i].BetAmount -= int(t.BetHight)
		if t.Players[i].BetAmount != 0 {
			sidePot += t.Players[i].BetAmount
		}

		if p.Bet == result {
			winners = append(winners, i)
		}
	}

	deductFromSidePot := 0

	if len(winners) == 0 {
		for i := range t.Players {
			t.Players[i].Wallet += (int(t.BetHight) + t.Players[i].BetAmount + minimumBet)
		}
		t.setForNextRound()
		return
	}

	if sidePot != 0 {
		for _, v := range winners {
			if t.Players[v].BetAmount != 0 {
				sidePotWin := (t.Players[v].BetAmount / sidePot) * 100
				t.Players[v].Wallet += sidePotWin
				deductFromSidePot += sidePotWin
			}
		}

		sidePot -= deductFromSidePot
		if sidePot > 0 {
			basePot += sidePot
		}
	}

	baseWinAmount := basePot / len(winners)

	for _, v := range winners {
		t.Players[v].Wallet += baseWinAmount
	}

	t.setForNextRound()
}

func (t *table) setForNextRound() {
	t.Point = 0
	t.BetHight = 0
	t.MaxBetHight = 0
	for i := range t.Players {
		t.Players[i].LastRoll[0] = 0
		t.Players[i].LastRoll[1] = 0
		t.Players[i].IsShooter = false
		t.Players[i].BetAmount = 0
		t.Players[i].Bet = ""
	}
}

func (t *table) letNonShootersBet() {
	for c := range t.wsConnections {
		if !c.player.IsShooter {
			c.send <- html.ShowWagerControlls(int(t.BetHight))
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
