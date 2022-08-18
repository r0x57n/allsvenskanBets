package main

import (
    "sort"
    "fmt"
    "strconv"
    "log"
    "time"
    _ "github.com/lib/pq"
    dg "github.com/bwmarrin/discordgo"
)

// Command: sammanfatta
func (b *botHolder) summaryCommand(i *dg.InteractionCreate) {
    if b.notOwner(getInteractUID(i)) { return }

    db := connectDB()
    defer db.Close()

    today := time.Now().Format("2006-01-02")

    round := -1
    err := db.QueryRow("SELECT round FROM matches WHERE date(date)>=$1 AND finished='0' ORDER BY date", today).Scan(&round)
    if err != nil { log.Panic(err) }

    var matches []match
    matchesRows, err := db.Query("SELECT id, hometeam, awayteam, date, homescore, awayscore, finished FROM matches WHERE round=$1", round)
    if err != nil { log.Panic(err) }

    won, lost := 0, 0
    err = db.QueryRow("SELECT COUNT(id) FROM bets WHERE round=$1 AND status=$2", round, BetStatusWon).Scan(&lost)
    if err != nil { log.Panic(err) }
    err = db.QueryRow("SELECT COUNT(id) FROM bets WHERE round=$1 AND status=$2", round, BetStatusLost).Scan(&won)
    if err != nil { log.Panic(err) }

    var bets []bet
    wins := make(map[int]int)

    for matchesRows.Next() {
        var m match
        matchesRows.Scan(&m.id, &m.hometeam, &m.awayteam, &m.date, &m.homescore, &m.awayscore, &m.finished)
        matches = append(matches, m)

        betsRows, err := db.Query("SELECT id, uid, matchid, homescore, awayscore, status FROM bets WHERE matchid=$1", m.id)
        if err != nil { log.Panic(err) }

        for betsRows.Next() {
            var b bet
            betsRows.Scan(&b.id, &b.uid, &b.matchid, &b.homescore, &b.awayscore, &b.status)
            bets = append(bets, b)

            if b.status == BetStatusWon {
                wins[b.uid]++
            }
        }
    }

    // Top three wins
    topThree := "Dom med flest vinster är:\n"
    keys := make([]int, 0, len(wins))
    for k := range wins {
        keys = append(keys, k)
    }

    sort.Ints(keys)

    for i, k := range keys {
        if i <= 3 {
            username, _ := b.session.User(strconv.Itoa(k))
            topThree += fmt.Sprintf("#%v - %v med %v vinster\n", i + 1, username.Username, wins[k])
        }
    }

    msg := fmt.Sprintf("Denna omgång spelades **%v** matcher och **%v** vadslagningar las.\n\n", len(matches), len(bets))
    msg += fmt.Sprintf("**%v**:st vann sina vad medans **%v**:st förlorade.\n\n", won, lost)
    msg += topThree

    addInteractionResponse(b.session, i, NewMsg, msg)
}
