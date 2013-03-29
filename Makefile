link_test: link_test.s
	gcc -o $@ $<
link_test.s: link_test.ll
# http://lists.cs.uiuc.edu/pipermail/llvmdev/2011-August/042128.html
	llc -disable-cfi -o $@ $<
link_test.ll: printnum.ll test.ll
	llvm-link printnum.ll test.ll -S -o link_test.ll
printnum.ll: printnum.c
	clang -emit-llvm -S -O -o $@ $<
test.ll: test.xxx
	go run code_gen.go $< $@
clean:
	rm -f *.ll *.s link_test
