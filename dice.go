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
	wsConnections map[*Client]bool
}

func (p *player) roll() {
	p.placeBet(10)
	p.LastRoll[0] = uint(rand.Intn(6-1) + 1)
	p.LastRoll[1] = uint(rand.Intn(6-1) + 1)
}

func (p *player) placeBet(amount int) {
	if p.Wallet-amount < 0 {
		return
	}
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
