BPF_CLANG = clang
BPF_CFLAGS = -O2 -g -Wall -target bpf -D__TARGET_ARCH_x86 -I$(PWD)/bpf

BPF_OBJS = \
	bpf/traffic.bpf.o \
	bpf/syscalls.bpf.o

all: $(BPF_OBJS)

bpf/%.bpf.o: bpf/%.bpf.c bpf/%.h bpf/vmlinux.h
	$(BPF_CLANG) $(BPF_CFLAGS) -c $< -o $@

clean:
	rm -f bpf/*.o
