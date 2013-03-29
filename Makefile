# link_test: link_test.s
# 	gcc -o link_test link_test.s
# link_test.s: link_test.ll
# 	llc -o link_test.s link_test.ll

link_test.ll: printnum.ll test.ll
	llvm-link test.ll printnum.ll -S -o  link_test.ll
printnum.ll: printnum.c
	clang -emit-llvm -S -O -o printnum.ll printnum.c
test.ll: test.xxx
	go run code_gen.go test.xxx 2> test.ll
clean:
	rm -f *.ll *.s
