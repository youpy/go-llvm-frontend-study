link_test: link_test.s
	gcc -o link_test link_test.s
link_test.s: link_test.ll
# http://lists.cs.uiuc.edu/pipermail/llvmdev/2011-August/042128.html
	llc -disable-cfi -o link_test.s link_test.ll
link_test.ll: printnum.ll test.ll
	llvm-link test.ll printnum.ll -S -o  link_test.ll
printnum.ll: printnum.c
	clang -emit-llvm -S -O -o printnum.ll printnum.c
test.ll: test.xxx
	go run code_gen.go test.xxx 2> test.ll
clean:
	rm -f *.ll *.s link_test
