.PHONY: dev test web-api web-test api-test

dev:
	@echo "Start the API and web app in separate terminals:"
	@echo "  cd apps/api && go run main.go api.go"
	@echo "  cd apps/web && npm install && npm run dev"

test: web-test api-test

web-test:
	cd apps/web && npm run typecheck && npm run lint && npm run test && npm run build

api-test:
	cd apps/api && go test ./...
