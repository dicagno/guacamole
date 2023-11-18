.PHONY: deploy

USER ?= user
IP ?= 192.168.1.1

all: build

build:
	@mkdir -p ./release
	go build -o ./release/guacamole main.go

run: build
	./release/guacamole

gen-proto:
	protoc -I proto/ proto/*.proto --go_out=packet

deploy: build
	@echo "Deploying to $(USER)@$(IP)"
	scp ./conf.d/guacamole.conf $(USER)@$(IP):/etc/guacamole.conf && \
	scp ./systemd/guacamole.service $(USER)@$(IP):/etc/systemd/system/ && \
	ssh $(USER)@$(IP) 'sudo systemctl enable guacamole && sudo systemctl start guacamole'

install: build
	@echo "Installing ..."
	sudo cp ./release/guacamole /usr/local/bin/ && \
	sudo cp ./conf.d/guacamole.conf /etc/guacamole.conf && \
	sudo cp ./systemd/guacamole.service /etc/systemd/system/ && \
    sudo systemctl enable guacamole && sudo systemctl start guacamole

