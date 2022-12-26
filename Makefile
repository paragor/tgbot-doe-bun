.PHONY: docker
docker:
	docker build . -t paragor/tgbot-doe-bun:latest
	docker push paragor/tgbot-doe-bun:latest
