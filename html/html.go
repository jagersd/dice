package html

import (
	"bytes"
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.html
var templateFiles embed.FS

//go:embed public
var StaticFiles embed.FS

var views = map[string]*template.Template{
	"home":  parse("home"),
	"game":  parse("game"),
	"lobby": parse("lobby"),
}

func Home(w io.Writer) error {
	return views["home"].Execute(w, "")
}

func Lobby(w io.Writer, data interface{}) error {
	return views["lobby"].Execute(w, data)
}

func LobbyPlayerList(w io.Writer, data []string) error {
	return views["lobby"].ExecuteTemplate(w, "playerList", data)
}

func Game(w io.Writer, data interface{}, index int) error {
	type enterGame struct {
		Table       interface{}
		PlayerIndex int
	}

	return views["game"].Execute(w, enterGame{
		Table:       data,
		PlayerIndex: index,
	})
}

func GameState(w io.Writer, data interface{}) error {
	return views["game"].ExecuteTemplate(w, "gameState", data)
}

func WSGameState(data interface{}) []byte {
	var buffer bytes.Buffer
	views["game"].ExecuteTemplate(&buffer, "gameState", data)
	return buffer.Bytes()
}

func Roll() []byte {
	var buffer bytes.Buffer
	views["game"].ExecuteTemplate(&buffer, "bet", nil)
	return buffer.Bytes()
}

func ShowActiveTables(w io.Writer, ts map[string]string) error {
	return views["home"].ExecuteTemplate(w, "activeTables", ts)
}

func ShowJoinForm(w io.Writer, data interface{}) error {
	return views["home"].ExecuteTemplate(w, "joinForm", data)
}

func parse(file string) *template.Template {
	file = "templates/" + file + ".html"
	return template.Must(
		template.New("layout.html").ParseFS(templateFiles, "templates/layout.html", file))
}
