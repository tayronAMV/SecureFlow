BPF_CLANG = clang
BPF_CFLAGS = -O2 -g -Wall -target bpf -D__TARGET_ARCH_x86 -I$(PWD)/bpf

BPF_OBJ = bpf/traffic.bpf.o

all: $(BPF_OBJ)

$(BPF_OBJ): bpf/traffic.bpf.c bpf/traffic.h bpf/vmlinux.h
	$(BPF_CLANG) $(BPF_CFLAGS) -c $< -o $@

clean:
	rm -f bpf/*.o
