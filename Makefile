migrate:
	$(MAKE) -C internal/auth migrate
	$(MAKE) -C internal/blog migrate

run-auth:
	$(MAKE) -C internal/auth run

run-blog:
	$(MAKE) -C internal/blog run

run-notification:
	$(MAKE) -C internal/notification run
