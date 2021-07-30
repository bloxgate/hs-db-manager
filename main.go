package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rivo/tview"
)

type Config struct {
	Username string
	Password string
	Protocol string
	Url      string `toml:"database_url"`
	Database string `toml:"database_name"`
}

var (
	app    *tview.Application
	pages  *tview.Pages
	db     *sql.DB
	config Config
)

func main() {
	app = tview.NewApplication()
	pages = tview.NewPages()

	var err error
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}

	//pages.AddPage("menu", makeMainMenu(), true, true)
	makeMainMenu()

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true", config.Username, config.Password, config.Protocol, config.Url, config.Database))
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	defer db.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("Error starting application: %v\n", err)
	}
}

func popupOverPage(f tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(f, height, 1, false).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}

func makeMainMenu() {
	menu := tview.NewList().ShowSecondaryText(false)
	menu.SetBorder(true).SetTitle("Main Menu")
	subMenu := tview.NewList().ShowSecondaryText(false)
	subMenu.SetBorder(true).SetTitle("Operations")
	subMenu.SetDoneFunc(func() {
		app.SetFocus(menu)
	})
	ioMenu := tview.NewForm()
	ioMenu.SetBorder(true).SetTitle("Data Input")
	ioMenu.SetCancelFunc(func() {
		app.SetFocus(subMenu)
	})

	flexLayout := tview.NewFlex().AddItem(menu, 0, 1, true).AddItem(subMenu, 0, 1, false).AddItem(ioMenu, 0, 3, false)

	menu.AddItem("Admins", "", 'a', func() {
		subMenu.Clear()
		subMenu.AddItem("Add", "", 'a', func() {
			ioMenu.Clear(true)
			ioMenu.AddInputField("ckey", "", 32, nil, nil)
			ioMenu.AddInputField("rank", "", 32, nil, nil)
			ioMenu.AddInputField("flags", "", 5, func(textToCheck string, lastChar rune) bool {
				_, err := strconv.ParseInt(textToCheck, 10, 16)
				return err == nil
			}, nil)
			ioMenu.AddButton("Execute", func() {
				ckeyInput := ioMenu.GetFormItemByLabel("ckey").(*tview.InputField).GetText()
				rankInput := ioMenu.GetFormItemByLabel("rank").(*tview.InputField).GetText()
				flagsString := ioMenu.GetFormItemByLabel("flags").(*tview.InputField).GetText()
				flagsNum, err := strconv.ParseInt(flagsString, 10, 16)
				if err != nil {
					container := tview.NewFlex().SetDirection(tview.FlexRow)
					container.SetTitle("ERROR").SetBorder(true)
					textView := tview.NewTextView().SetDynamicColors(true).SetText("[red]Flags must be a number between 0 and 65356.").SetDoneFunc(func(_ tcell.Key) {
						pages.RemovePage("updateResults")
						app.SetFocus(ioMenu)
					})
					container.AddItem(textView, 0, 1, true)
					pages.AddPage("updateResults", popupOverPage(container, 40, 10), true, true)
					app.SetFocus(textView)
				} else {
					pages.AddPage("updateResults", addToAdmins(ckeyInput, rankInput, uint16(flagsNum), ioMenu), true, true)
				}
			})
			app.SetFocus(ioMenu)
		})
		subMenu.AddItem("Remove", "", 'r', func() {
			ioMenu.Clear(true)
			ioMenu.AddButton("Not Implemented. Perform in game.", nil)
		})
		subMenu.AddItem("Search", "", 's', func() {
			ioMenu.Clear(true)
			ioMenu.AddDropDown("Type", []string{"ckey", "rank"}, 0, nil).
				AddInputField("Search Term", "", 32, nil, nil).
				AddButton("Search", func() {
					typeInput, _ := ioMenu.GetFormItemByLabel("Type").(*tview.DropDown).GetCurrentOption()
					searchTerm := ioMenu.GetFormItemByLabel("Search Term").(*tview.InputField).GetText()
					pages.AddPage("searchResults", searchForAdmins(typeInput == 0, searchTerm, ioMenu), true, true)
				})
			app.SetFocus(ioMenu)
		})
		subMenu.AddItem("Update", "", 'u', func() {
			ioMenu.Clear(true)
			ioMenu.AddButton("Not Implemented. Perform in game.", nil)
		})
		app.SetFocus(subMenu)
	})
	menu.AddItem("Bans", "", 'b', func() {
		subMenu.Clear()
		subMenu.AddItem("Remove", "", 'r', func() {
			ioMenu.Clear(true)
			ioMenu.AddButton("Not Implemented. Unban in game.", nil)
		})
		subMenu.AddItem("Search", "", 's', func() {
			ioMenu.Clear(true)
			ioMenu.AddDropDown("Search By", []string{"ckey", "computerid", "ip", "a_ckey"}, 0, nil)
			ioMenu.AddInputField("Search Term", "", 32, nil, nil)
			ioMenu.AddButton("Search", func() {
				_, typeInput := ioMenu.GetFormItemByLabel("Search By").(*tview.DropDown).GetCurrentOption()
				searchTerm := ioMenu.GetFormItemByLabel("Search Term").(*tview.InputField).GetText()
				pages.AddPage("searchResults", searchForBans(typeInput, searchTerm, ioMenu), true, true)
			})
			app.SetFocus(ioMenu)
		})
		app.SetFocus(subMenu)
	})
	menu.AddItem("Whitelist", "", 'w', func() {
		subMenu.Clear()
		subMenu.AddItem("Add", "", 'a', func() {
			ioMenu.Clear(true)
			ioMenu.AddInputField("ckey", "", 32, nil, nil)
			ioMenu.AddInputField("race", "", 32, nil, nil)
			ioMenu.AddButton("Execute", func() {
				ckeyInput := ioMenu.GetFormItemByLabel("ckey").(*tview.InputField).GetText()
				raceInput := ioMenu.GetFormItemByLabel("race").(*tview.InputField).GetText()
				pages.AddPage("updateResults", addToWhitelist(ckeyInput, raceInput, ioMenu), true, true)
			})
			app.SetFocus(ioMenu)
		})
		subMenu.AddItem("Remove", "", 'r', func() {
			ioMenu.Clear(true)
			ioMenu.AddInputField("ID", "", 8, nil, nil)
			ioMenu.AddButton("Execute", func() {
				modal := tview.NewModal().SetText("WARNING. This operation is destructive and can not be undone. Are you sure you want to continue?").SetTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorRed).
					AddButtons([]string{"No", "Yes"}).
					SetDoneFunc(func(idx int, label string) {
						app.SetRoot(pages, true)
						if label == "No" || label == "" {
							app.SetFocus(ioMenu)
						} else {
							id := ioMenu.GetFormItemByLabel("ID").(*tview.InputField).GetText()
							idNum, err := strconv.ParseInt(id, 10, 16)
							if err != nil {
								container := tview.NewFlex().SetDirection(tview.FlexRow)
								container.SetTitle("ERROR").SetBorder(true)
								textView := tview.NewTextView().SetDynamicColors(true).SetText("[red]ID must be numeric").SetDoneFunc(func(_ tcell.Key) {
									pages.RemovePage("updateResults")
									app.SetFocus(ioMenu)
								})
								container.AddItem(textView, 0, 1, true)
								pages.AddPage("updateResults", popupOverPage(container, 40, 10), true, true)
								app.SetFocus(textView)
							} else {
								pages.AddPage("updateResults", removeFromWhitelist(int16(idNum), ioMenu), true, true)
							}
						}
					})
				app.SetRoot(modal, false)
				app.SetFocus(modal)
			})
			app.SetFocus(ioMenu)
		})
		subMenu.AddItem("Search", "", 's', func() {
			ioMenu.Clear(true)
			ioMenu.AddDropDown("Search By", []string{"ckey", "race"}, 0, nil)
			ioMenu.AddInputField("Search Term", "", 32, nil, nil)
			ioMenu.AddButton("Search", func() {
				_, typeInput := ioMenu.GetFormItemByLabel("Search By").(*tview.DropDown).GetCurrentOption()
				searchTerm := ioMenu.GetFormItemByLabel("Search Term").(*tview.InputField).GetText()
				pages.AddPage("searchResults", searchForWhiteList(typeInput, searchTerm, ioMenu), true, true)
			})
			app.SetFocus(ioMenu)
		})
		app.SetFocus(subMenu)
	})
	menu.AddItem("Quit", "", 'q', func() {
		app.Stop()
	})

	pages.AddPage("menu", flexLayout, true, true)
	app.SetRoot(pages, true)
}

func addToWhitelist(ckey string, race string, returnFocus tview.Primitive) tview.Primitive {
	stmt, err := db.Prepare("INSERT INTO whitelist(ckey, race) VALUES (?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(ckey, race)
	if err != nil {
		panic(err)
	}

	newId, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	textView := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf("[green]Inserted %d row(s). Last inserted ID: %d", rowsUpdated, newId)).SetDoneFunc(func(_ tcell.Key) {
		pages.RemovePage("updateResults")
		app.SetFocus(returnFocus)
	})

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Execution Results").SetBorder(true)
	container.AddItem(textView, 0, 1, true)

	app.SetFocus(textView)

	return popupOverPage(container, 40, 10)
}

func removeFromWhitelist(id int16, returnFocus tview.Primitive) tview.Primitive {
	stmt, err := db.Prepare("DELETE FROM whitelist WHERE id = ? LIMIT 1")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		panic(err)
	}

	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	textView := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf("[yellow]Updated %d rows.\n[red]If this is greater than 1, something went very wrong.", rowsUpdated)).SetDoneFunc(func(_ tcell.Key) {
		pages.RemovePage("updateResults")
		app.SetFocus(returnFocus)
	})

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Execution Results").SetBorder(true)
	container.AddItem(textView, 0, 1, true)

	app.SetFocus(textView)

	return popupOverPage(container, 40, 10)
}

func searchForWhiteList(searchType string, searchTerm string, returnFocus tview.Primitive) tview.Primitive {
	var rows *sql.Rows
	var err error

	rows, err = db.Query(fmt.Sprintf("SELECT id,ckey,race FROM whitelist WHERE %s LIKE ?", searchType), searchTerm)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Search Results").SetBorder(true)
	table := tview.NewTable()

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	for idx, name := range columns {
		table.SetCell(0, idx, tview.NewTableCell(name).SetTextColor(tcell.ColorRed))
	}

	dest := struct {
		id   int16
		ckey string
		race string
	}{}
	for rows.Next() {
		rows.Scan(&dest.id, &dest.ckey, &dest.race)

		row := table.GetRowCount()
		table.SetCellSimple(row, 0, fmt.Sprintf("%d", dest.id))
		table.SetCellSimple(row, 1, dest.ckey)
		table.SetCellSimple(row, 2, dest.race)
	}
	//Check for an error getting the next row
	if err = rows.Err(); err != nil {
		panic(err)
	}

	container.AddItem(table, 0, 1, true)

	table.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("searchResults")
		app.SetFocus(returnFocus)
	})
	app.SetFocus(table)

	return popupOverPage(container, 40, 10)
}

func searchForBans(searchType string, searchTerm string, returnFocus tview.Primitive) tview.Primitive {
	var rows *sql.Rows
	var err error

	rows, err = db.Query(fmt.Sprintf("SELECT bantime,bantype,reason,job,duration,ckey,computerid,ip,a_ckey FROM erro_ban WHERE %s LIKE ?", searchType), searchTerm)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Search Results").SetBorder(true)
	table := tview.NewTable()

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	for idx, name := range columns {
		table.SetCell(0, idx, tview.NewTableCell(name).SetTextColor(tcell.ColorRed))
	}

	dest := struct {
		bantime    time.Time
		bantype    string
		reason     string
		job        string
		duration   int16
		ckey       string
		computerid string
		ip         string
		a_ckey     string
	}{}
	for rows.Next() {
		rows.Scan(&dest.bantime, &dest.bantype, &dest.reason, &dest.job, &dest.duration, &dest.ckey, &dest.computerid, &dest.ip, &dest.a_ckey)

		row := table.GetRowCount()
		table.SetCellSimple(row, 0, dest.bantime.Format(time.RFC3339))
		table.SetCellSimple(row, 1, dest.bantype)
		table.SetCellSimple(row, 2, dest.reason)
		table.SetCellSimple(row, 3, dest.job)
		table.SetCellSimple(row, 4, fmt.Sprintf("%d", dest.duration))
		table.SetCellSimple(row, 5, dest.ckey)
		table.SetCellSimple(row, 6, dest.computerid)
		table.SetCellSimple(row, 7, dest.ip)
		table.SetCellSimple(row, 8, dest.a_ckey)
	}
	//Check for an error getting the next row
	if err = rows.Err(); err != nil {
		panic(err)
	}

	container.AddItem(table, 0, 1, true)

	table.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("searchResults")
		app.SetFocus(returnFocus)
	})
	app.SetFocus(table)

	return popupOverPage(container, 80, 24)
}

func addToAdmins(ckey string, rank string, flags uint16, returnFocus tview.Primitive) tview.Primitive {
	stmt, err := db.Prepare("INSERT INTO erro_admin(ckey, rank, level, flags) VALUES (?, ?, -1, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(ckey, rank, flags)
	if err != nil {
		panic(err)
	}

	newId, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	textView := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf("[green]Inserted %d row(s). Last inserted ID: %d", rowsUpdated, newId)).SetDoneFunc(func(_ tcell.Key) {
		pages.RemovePage("updateResults")
		app.SetFocus(returnFocus)
	})

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Execution Results").SetBorder(true)
	container.AddItem(textView, 0, 1, true)

	app.SetFocus(textView)

	return popupOverPage(container, 40, 10)
}

func searchForAdmins(ckey bool, searchTerm string, returnFocus tview.Primitive) tview.Primitive {
	var rows *sql.Rows
	var err error
	if ckey {
		rows, err = db.Query("SELECT * FROM erro_admin WHERE ckey LIKE ?", searchTerm)
		if err != nil {
			panic(err)
		}
	} else {
		rows, err = db.Query("SELECT * FROM erro_admin WHERE rank LIKE ?", searchTerm)
		if err != nil {
			panic(err)
		}
	}
	defer rows.Close()

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Search Results").SetBorder(true)
	table := tview.NewTable()
	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	for idx, name := range columns {
		table.SetCell(0, idx, tview.NewTableCell(name).SetTextColor(tcell.ColorRed))
	}

	dest := struct {
		id    int16
		ckey  string
		rank  string
		level int8
		flags uint16
	}{}
	for rows.Next() {
		rows.Scan(&dest.id, &dest.ckey, &dest.rank, &dest.level, &dest.flags)

		//Get current row
		row := table.GetRowCount()
		table.SetCellSimple(row, 0, fmt.Sprintf("%d", dest.id))
		table.SetCellSimple(row, 1, dest.ckey)
		table.SetCellSimple(row, 2, dest.rank)
		table.SetCellSimple(row, 3, fmt.Sprintf("%d", dest.level))
		table.SetCellSimple(row, 4, fmt.Sprintf("%d", dest.flags))
	}
	//Check for an error getting the next row
	if err = rows.Err(); err != nil {
		panic(err)
	}

	container.AddItem(table, 0, 1, true)

	table.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("searchResults")
		app.SetFocus(returnFocus)
	})
	app.SetFocus(table)

	return popupOverPage(container, 40, 10)
}
