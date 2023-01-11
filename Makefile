.PHONY: docker
docker:
	docker buildx build --push --platform linux/amd64,linux/arm64 . -t paragor/tgbot-doe-bun:latest
