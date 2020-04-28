# AMA Bot

A Discord bot made specifically for hosting a single AMA.

## How to use

1. Prepare the two environment variables:
  - The owner's Discord user ID as `OWNER`.
  - The bot's Discord token as `TOKEN`.

2. Run the bot with 
```bash
go run cmd/bot/main.go
```

3. DM the bot:
```
ama!init [server ID] [ask channel] [main channel]
```

4. Follow the bot's instructions.

