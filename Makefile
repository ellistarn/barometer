

verify: ## Verify code. Includes codegen, dependencies, linting, formatting, etc
	go mod tidy
	go generate ./...
	go vet ./...
	golangci-lint run
	@git diff --quiet ||\
		{ echo "New file modification detected in the Git working tree. Please check in before commit."; git --no-pager diff --name-only | uniq | awk '{print "  - " $$0}'; \
		if [ "${CI}" == 'true' ]; then\
			exit 1;\
		fi;}

apply:
	helm template chart | ko apply -B -f -

delete:
	helm template chart | ko delete -f -

example:
	ko apply -B -f examples
