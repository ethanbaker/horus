module github.com/ethanbaker/horus/bot

replace github.com/ethanbaker/horus/utils => ../utils

go 1.20

require (
	github.com/ethanbaker/horus/utils v0.0.0-00010101000000-000000000000
	github.com/sashabaranov/go-openai v1.22.0
	github.com/stretchr/objx v0.5.2
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.10
)

require (
	github.com/bwmarrin/discordgo v0.28.1 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)
