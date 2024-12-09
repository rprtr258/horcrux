.PHONY: test-diary
test-diary:
	rm diary.txt.*.horcrux output.txt || true
	go run . split -n 4 -k 3 diary.txt
	# go run . join diary.txt.{0,1,3}.horcrux
	go run . join diary.txt.0.horcrux diary.txt.1.horcrux diary.txt.3.horcrux
	cmp output.txt diary.txt
