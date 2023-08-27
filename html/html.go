package html

import (
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

func ShowActiveTables(w io.Writer, ts map[string]string) error {
	type table struct {
		Slug string
		Name string
	}
	var tables []table
	for s, n := range ts {
		tables = append(tables, table{
			Slug: s,
			Name: n,
		})
	}
	return views["home"].ExecuteTemplate(w, "activeTables", tables)
}

func parse(file string) *template.Template {
	file = "templates/" + file + ".html"
	return template.Must(
		template.New("layout.html").ParseFS(templateFiles, "templates/layout.html", file))
}
