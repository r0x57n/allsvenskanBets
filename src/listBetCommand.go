package main

import (
	"fmt"
	"strconv"
	"log"
	_ "github.com/mattn/go-sqlite3"
	dg "github.com/bwmarrin/discordgo"
)

// Command: vadslagningar
func listBetsCommand(s *dg.Session, i *dg.InteractionCreate) {
    db := connectDB()
	defer db.Close()

    // Get options and parse
    options := getOptionsOrRespond(s, i, NewMsg)
    if options == nil { return }
	uid := options[0].Value

    listTypes := All
    if len(options) == 2 {
        listTypes, _ = strconv.Atoi(fmt.Sprintf("%v", options[1].Value))
    }

    desc := ""

    userToView := getUser(db, fmt.Sprint(uid))
    userNotViewingThemselves := userToView.uid != getUserFromInteraction(db, i).uid
    if userToView.viewable == 0 && userNotViewingThemselves {
		desc := "Användaren har valt att dölja sina vadslagningar."
        addInteractionResponse(s, i, NewMsg, desc)
        return
    } else if userToView.viewable == 0 && !userNotViewingThemselves {
        desc = "Andra användare kan inte se dina vadslagningar."
    }

    where := ""
    switch listTypes {
        case Lost:
            where = "uid=? AND handled=1 AND won=0"
        case Won:
            where = "uid=? AND handled=1 AND won=1"
        case All:
            where = "uid=? AND handled=1"
        default:
            addErrorResponse(s, i, NewMsg, "Got unidentifiable listTypes in listBetsCommand.")
            return
    }

    rows, err := db.Query("SELECT b.homeScore, b.awayScore, b.matchid, b.won, m.homeTeam, m.awayTeam " +
                          "FROM bets AS b " +
                          "JOIN matches AS m ON b.matchid=m.id " +
                          "WHERE " + where, uid)
    defer rows.Close()
    if err != nil { log.Panic(err) }

    wonBets, lostBets := "-", "-"

    for rows.Next() {
        var b bet
        var m match

        rows.Scan(&b.homeScore, &b.awayScore, &b.matchid, &b.won, &m.homeTeam, &m.awayTeam)

        switch b.won {
            case 0:
                if lostBets == "-" { lostBets = "" }
                lostBets += fmt.Sprintf("%v (**%v**) - %v (**%v**)\n", m.homeTeam, b.homeScore, m.awayTeam, b.awayScore)
            case 1:
                if wonBets == "-" { wonBets = "" }
                wonBets += fmt.Sprintf("%v (**%v**) - %v (**%v**)\n", m.homeTeam, b.homeScore, m.awayTeam, b.awayScore)
            default:
                addErrorResponse(s, i, NewMsg, "Got unidentifiable b.won in listBetsCommand.")
                return
        }
    }

    if wonBets == "-" && lostBets == "-" {
        desc = fmt.Sprintf("Användaren har inga vadslagningar ännu!", )
    }

    rows, err = db.Query("SELECT c.challengerUID, c.challengeeUID, c.points, c.condition, " +
                         "m.homeTeam, m.awayTeam " +
                         "FROM challenges AS c " +
                         "JOIN matches AS m ON c.matchID=m.id " +
                         "WHERE (c.challengerUID=? OR c.challengeeUID=?) AND c.status=?", uid, uid, Handled)
    defer rows.Close()
    if err != nil { log.Panic(err) }

    type challAndMatch struct {
        c challenge
        m match
    }

    challenges := "-"

    for rows.Next() {
        var c challenge
        var m match

        rows.Scan(&c.challengerUID, &c.challengeeUID, &c.points, &c.condition,
                  &m.homeTeam, &m.awayTeam)

        challenger, err := s.User(strconv.Itoa(c.challengerUID))
        if err != nil { log.Panic(err) }
        challengee, err := s.User(strconv.Itoa(c.challengeeUID))
        if err != nil { log.Panic(err) }

        winOrLose := "vinna"
        if c.condition == "awayTeam" {
            winOrLose = "förlora"
        }

        if challenges == "-" { challenges = "" }
        challenges += fmt.Sprintf("**%v** utmanade **%v** om att **%v** skulle %v mot **%v** för **%v** poäng\n",
                                    challenger.Username, challengee.Username, m.homeTeam, winOrLose, m.awayTeam, c.points)
    }

    fields := []*dg.MessageEmbedField {}

    if listTypes == 0 {
        fields = []*dg.MessageEmbedField {
            {
                Name: "Förlorade",
                Value: lostBets,
            },
        }
    } else if listTypes == 1 {
        fields = []*dg.MessageEmbedField {
            {
                Name: "Vunna",
                Value: wonBets,
            },
        }
    } else {
        fields = []*dg.MessageEmbedField {
            {
                Name: "Vunna",
                Value: wonBets,
            },
            {
                Name: "Förlorade",
                Value: lostBets,
            },
        }
    }

    fields = append(fields, &dg.MessageEmbedField{
        Name: "Utmaningar",
        Value: challenges,
    })

    addEmbeddedInteractionResponse(s, i, NewMsg, fields, "Vadslagningar", desc)
}
