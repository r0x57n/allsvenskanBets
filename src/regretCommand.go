package main

import (
    "fmt"
    "strconv"
    "log"
    "time"
    _ "github.com/mattn/go-sqlite3"
    dg "github.com/bwmarrin/discordgo"
)

func regretCommand(s *dg.Session, i *dg.InteractionCreate) {
    db := connectDB()
    defer db.Close()

    uid := getInteractUID(i)

    allBets := *getBets(db, "uid=? AND handled=?", uid, 0)

    labels := make(map[int]string)
    dates := make(map[int]string)

    var regrettableBets []bet

    for _, b := range allBets {
        m := getMatch(db, "id=?", b.matchid)
        matchDate, _ := time.Parse(DB_TIME_LAYOUT, m.date)

        if time.Now().Before(matchDate) {
            labels[b.id] = fmt.Sprintf("%v vs %v [%v-%v]", m.homeTeam, m.awayTeam, b.homeScore, b.awayScore)
            dates[b.id] = matchDate.Format(MSG_TIME_LAYOUT)
            regrettableBets = append(regrettableBets, b)
        }
    }

    if len(regrettableBets) == 0 {
        addInteractionResponse(s, i, NewMsg, "Inga framtida vadslagningar...")
        return
    }

    options := []dg.SelectMenuOption{}

    for _, b := range regrettableBets {
        options = append(options, dg.SelectMenuOption{
            Label: labels[b.id],
            Value: strconv.Itoa(b.id),
            Description: dates[b.id],
        })
    }

    components := []dg.MessageComponent {
        dg.ActionsRow {
            Components: []dg.MessageComponent {
                dg.SelectMenu {
                    Placeholder: "Välj en vadslagning",
                    CustomID: "regretSelected",
                    Options: options,
                },
            },
        },
    }

    addCompInteractionResponse(s, i, NewMsg, "Dina vadslagningar", components)
}

func regretSelected(s *dg.Session, i *dg.InteractionCreate) {
    db := connectDB()
    defer db.Close()

    values := getValuesOrRespond(s, i, UpdateMsg)
    if values == nil { return }
    bid := values[0]

    var m match
    var b bet
    err := db.QueryRow("SELECT m.date, b.uid FROM bets AS b " +
                       "JOIN matches AS m ON b.matchid=m.id " +
                       "WHERE b.id=?", bid).Scan(&m.date, &b.uid)
    if err != nil { log.Panic(err) }

    components := []dg.MessageComponent {}
    msg := ""

    if matchHasBegun(s, i, m) {
        msg = "Kan inte ta bort en vadslagning om en pågående match..."
    } else {
        if strconv.Itoa(b.uid) != getInteractUID(i) {
            addErrorResponse(s, i, UpdateMsg, "Du försökte ta bort någon annans vadslagning...")
            return
        }

        _, err = db.Exec("DELETE FROM bets WHERE id=?", bid)
        msg = "Vadslagning borttagen!"
    }

    addCompInteractionResponse(s, i, UpdateMsg, msg, components)
}
