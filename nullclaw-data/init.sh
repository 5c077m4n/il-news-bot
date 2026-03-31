apk add --no-cache jq

jq --arg TELEGRAM_API_KEY "$TELEGRAM_API_KEY" '.channels.telegram.accounts.default.bot_token = $TELEGRAM_API_KEY' /nullclaw-data/config.json >/nullclaw-data/config.json
jq --arg TELEGRAM_USER_ID "$TELEGRAM_USER_ID" '.channels.telegram.accounts.default.allow_from = $TELEGRAM_USER_ID' /nullclaw-data/config.json >/nullclaw-data/config.json
jq --arg OPENROUTER_API_KEY "$OPENROUTER_API_KEY" '.models.providers.openrouter.api_key = $OPENROUTER_API_KEY' /nullclaw-data/config.json >/nullclaw-data/config.json

nullclaw service install
nullclaw service start
nullclaw channel start telegram
nullclaw gateway
