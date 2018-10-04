bench:
	go test -bench=. | column -t

test:
	go test -v -cover
