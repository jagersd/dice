package main

import (
	"errors"
	"math/rand"
)

var activeTables = make(map[string]*table, 10)

type player struct {
	Name      string
	BetAmount int
	Bet       string
	Wallet    int
	LastRoll  []uint
}

type table struct {
	Name         string
	InternalName string
	Players      []player
	Point        uint
	Complete     bool
}

func (p *player) roll() {
	p.LastRoll[0] = uint(rand.Intn(6-1) + 1)
	p.LastRoll[1] = uint(rand.Intn(6-1) + 1)
}

func (p *player) placeBet(amount int) {
	if p.Wallet-amount < 0 {
		return
	}
	p.Wallet -= amount
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

	activeTables[t.InternalName] = &t

	return t.InternalName, nil
}

func newPlayer(playerName string) player {
	return player{
		Name:   playerName,
		Wallet: 100,
	}
}
