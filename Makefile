build:
	mkdir -p bin/
	go build -o bin/leds .

run:
	bin/leds

clean:
	rm -f bin/leds

ci:
	make clean && make build && make run
