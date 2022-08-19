package main

import (
	"log"
	dg "github.com/bwmarrin/discordgo"
)

func (b *botHolder) Init() {
	log.Print("Initializing...")

	// Login bot to get the active session
	s, err := dg.New("Bot " + b.token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

    b.session = s

    b.addCommands()

    // Add handlers for command/component
    b.commandHandlers = map[string]func(s *dg.Session, i *dg.InteractionCreate) {
        // User commands
        HelpCommand:        func(s *dg.Session, i *dg.InteractionCreate) {    b.helpCommand(i)        },
        BetCommand:         func(s *dg.Session, i *dg.InteractionCreate) {    b.betCommand(i)         },
        RegretCommand:      func(s *dg.Session, i *dg.InteractionCreate) {    b.regretCommand(i)      },
        ChallengeCommand:   func(s *dg.Session, i *dg.InteractionCreate) {    b.challengeCommand(i)   },
        ChickenCommand:     func(s *dg.Session, i *dg.InteractionCreate) {    b.chickenCommand(i)     },
        UpcomingCommand:    func(s *dg.Session, i *dg.InteractionCreate) {    b.upcomingCommand(i)    },
        BetsCommand:        func(s *dg.Session, i *dg.InteractionCreate) {    b.listBetsCommand(i)    },
        PointsCommand:      func(s *dg.Session, i *dg.InteractionCreate) {    b.pointsCommand(i)      },
        SettingsCommand:    func(s *dg.Session, i *dg.InteractionCreate) {    b.settingsCommand(i)    },
        InfoCommand:        func(s *dg.Session, i *dg.InteractionCreate) {    b.infoCommand(i)        },

        // Admin commands
        SummaryCommand:     func(s *dg.Session, i *dg.InteractionCreate) {    b.summaryCommand(i)     },
        UpdateCommand:      func(s *dg.Session, i *dg.InteractionCreate) {    b.updateCommand(i)      },
        DeleteCommand:      func(s *dg.Session, i *dg.InteractionCreate) {    b.deleteCommand(i)      },
        CheckCommand:       func(s *dg.Session, i *dg.InteractionCreate) {    b.checkBetsCommand(i)   },
    }

    // Component handlers
    b.componentHandlers = map[string]func(s *dg.Session, i *dg.InteractionCreate) {
        BetOnSelected:        func(s *dg.Session, i *dg.InteractionCreate) {   b.betOnSelected(i)              },
        BetScoreHome:         func(s *dg.Session, i *dg.InteractionCreate) {   b.betScoreComponent(i, Home)    },
        BetScoreAway:         func(s *dg.Session, i *dg.InteractionCreate) {   b.betScoreComponent(i, Away)    },
        ChallSelectWinner:    func(s *dg.Session, i *dg.InteractionCreate) {   b.challSelectWinner(i)          },
        ChallSelectPoints:    func(s *dg.Session, i *dg.InteractionCreate) {   b.challSelectPoints(i)          },
        ChallAcceptDiscard:   func(s *dg.Session, i *dg.InteractionCreate) {   b.challAcceptDiscard(i)         },
        ChallAcceptDiscardDo: func(s *dg.Session, i *dg.InteractionCreate) {   b.challAcceptDiscardDo(i)       },
        ChallAnswer:          func(s *dg.Session, i *dg.InteractionCreate) {   b.challAnswer(i)                },
        SettingsVisibility:   func(s *dg.Session, i *dg.InteractionCreate) {   b.settingsVisibility(i)         },
        SettingsChall:        func(s *dg.Session, i *dg.InteractionCreate) {   b.settingsChall(i)              },
        UpdateCommandDo:      func(s *dg.Session, i *dg.InteractionCreate) {   b.updateCommandDo(s, i)          },
        DeleteCommandDo:      func(s *dg.Session, i *dg.InteractionCreate) {   b.deleteCommandDo(s, i)          },
        RegretSelected:       func(s *dg.Session, i *dg.InteractionCreate) {   b.regretSelected(i)             },
        ChickenSelected:      func(s *dg.Session, i *dg.InteractionCreate) {   b.chickenSelected(i)            },
        ChickenAnswer:        func(s *dg.Session, i *dg.InteractionCreate) {   b.chickenAnswer(i)              },
    }

	s.AddHandler(func(s *dg.Session, i *dg.InteractionCreate) {
		switch i.Type {
			case dg.InteractionApplicationCommand:
                if h, ok := b.commandHandlers[i.ApplicationCommandData().Name]; ok { h(s, i) }
			case dg.InteractionMessageComponent:
				if h, ok := b.componentHandlers[i.MessageComponentData().CustomID]; ok { h(s, i) }
		}
	})

    // Handler to tell us when we logged in
	s.AddHandler(func(s *dg.Session, r *dg.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
}

func (b *botHolder) Start() {
	err := b.session.Open()
	if err != nil {
		log.Panicf("Cannot open the session: %v", err)
	}
}

func (b *botHolder) Close() {
    b.session.Close()
    b.db.Close()
}

func (b *botHolder) notOwner(uid string) bool {
    if b.owner != uid { return true }
    return false
}

func (b *botHolder) addCommands() {
    b.commands = []command {
        {
            name: HelpCommand,
            description: "Få hjälp med hur denna bot fungerar.",
            category: CommandCategoryGeneral,
        },
        {
            name: BetCommand,
            description: "Slå vad om en kommande match.",
            category: CommandCategoryBetting,
        },
        {
            name: RegretCommand,
            description: "Ångra ett vad du har lagt.",
            category: CommandCategoryBetting,
        },
        {
            name: ChallengeCommand,
            description: "Utmana en annan användare.",
            category: CommandCategoryBetting,
            options: []*dg.ApplicationCommandOption {
                {
                    Type: dg.ApplicationCommandOptionUser,
                    Name: "användarnamn",
                    Description: "Användare att utmana.",
                    Required: true,
                },
                {
                    Type: dg.ApplicationCommandOptionString,
                    Name: "typ",
                    Description: "Vilken sorts utmaning?",
                    Required: true,
                    Choices: []*dg.ApplicationCommandOptionChoice {
                        {
                            Name: "Matchvinnare",
                            Value: "matchvinnare",
                        },
                    },
                },
            },
        },
        {
            name: ChickenCommand,
            description: "Be om att få fega ur en utmaning.",
            category: CommandCategoryBetting,
        },
        {
            name: UpcomingCommand,
            description: "Lista dina kommande vadslagningar.",
            category: CommandCategoryListing,
        },
        {
            name: BetsCommand,
            description: "Lista en annan användares tidigare vadslagningar.",
            category: CommandCategoryListing,
            options: []*dg.ApplicationCommandOption {
                {
                    Type: dg.ApplicationCommandOptionUser,
                    Name: "användarnamn",
                    Description: "Användare att visa vadslagningar för.",
                    Required: true,
                },
                {
                    Type: dg.ApplicationCommandOptionString,
                    Name: "typ",
                    Description: "Vill du enbart visa en viss typ av vad?",
                    Choices: []*dg.ApplicationCommandOptionChoice {
                        {
                            Name: "vunna",
                            Value: "1",
                        },
                        {
                            Name: "förlorade",
                            Value: "0",
                        },
                    },
                },
            },
        },
        {
            name: PointsCommand,
            description: "Visa dina poäng och topp 10 på servern.",
            category: CommandCategoryListing,
        },
        {
            name: SettingsCommand,
            description: "Inställningar för din användare.",
            category: CommandCategoryGeneral,
        },
        {
            name: InfoCommand,
            description: "Teknisk info om botten.",
            category: CommandCategoryGeneral,
        },
        {
            name: SummaryCommand,
            description: "Sammanfatta denna omgång till #bets.",
            category: CommandCategoryAdmin,
            admin: true,
        },
        {
            name: UpdateCommand,
            description: "Uppdatera alla kommandon eller ett enskilt.",
            category: CommandCategoryAdmin,
            admin: true,
        },
        {
            name: DeleteCommand,
            description: "Ta bort ett enskilt kommando.",
            category: CommandCategoryAdmin,
            admin: true,
        },
        {
            name: CheckCommand,
            description: "Kör checks för challenges/bets.",
            category: CommandCategoryAdmin,
            admin: true,
        },
    }
}
