Телеграм бот в которого можно кинуть сообщением и он будет тебе его переотправлять каждый час пока ты его не закончишь

```bash
SECRET_KEY=supersecretkeytoauth TELEGRAM_APITOKEN=youknow BADGER_DB=$(pwd)/badger.db go run main.go
```

```text
> supersecretkeytoauth
> text
< ok, notify after 2022-12-27 03:00:27.33754 +0700 +07 m=+3609.708126334
< ^ 0
> done
```
