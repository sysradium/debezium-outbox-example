loadtest:
	ab  -p body.json -T "application/json" -n 100 -c10 -m POST http://127.0.0.1:8081/users

build-images: $(foreach svc,notifications-service users-service,build-image-$(svc))

build-image-%:
	docker build -t outbox-$* --build-arg SERVICE=$* .
