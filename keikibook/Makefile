.PHONY: install run build preview

node_modules: package.json package-lock.json
	npm install
	@touch node_modules

install: node_modules

run: node_modules
	npm run dev

build: node_modules
	npm run build

preview: node_modules
	npm run preview
