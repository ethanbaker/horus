module github.com/ethanbaker/horus/implementations/discord

replace github.com/ethanbaker/horus/bot => ../../bot

replace github.com/ethanbaker/horus/utils => ../../utils

replace github.com/ethanbaker/horus/outreach => ../../outreach

go 1.20

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/ethanbaker/horus/bot v0.0.0-00010101000000-000000000000
	github.com/ethanbaker/horus/outreach v0.0.0-20240517162421-da5329774036
	github.com/ethanbaker/horus/utils v0.0.0-20240419205637-d49093486dd8
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/sashabaranov/go-openai v1.22.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/arran4/golang-ical v0.2.8 // indirect
	github.com/dstotijn/go-notion v0.11.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/teambition/rrule-go v1.8.2 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/gorm v1.25.10 // indirect
)
