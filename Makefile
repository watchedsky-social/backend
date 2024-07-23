DAGGER := $(shell command -v dagger)

.PHONY: get-frontend
get-frontend: frontend/index.html

frontend/index.html:
	$(DAGGER) -m github.com/watchedsky-social/frontend \
		call get-built-site --registry=registry.lab.verysmart.house \
		--image-version=latest export --path=./frontend
