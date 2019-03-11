set disassembly-flavor intel
set confirm off
set pagination off

b ethereumUseGas
b main.exec
b ethereumFinish
b ethereumCallDataCopy
#b github.com/tsavola/wag/internal/isa/x86.MacroAssembler.CurrentMemory
b github.com/tsavola/wag/internal/isa/x86.MacroAssembler.GrowMemory

define ds
x/32x $rsp
end

define di
x/8i $rip
end

command 1
set language c
b  *((*(uint64_t *)$rsp & 0xffffff000000) + 0x16e00)
b  *((*(uint64_t *)$rsp & 0xffffff000000) + 0x16670)
set language auto

 command 6
 #printf "** Before return from allocate**\n"
 #set $rsp += 0x68
 #set language c
 #set $rcx = *(long long int *)($rsp-0x10)
 #set language auto
 #printf "%lx\n", $rcx
 #c
 end

c
end

command 5
c
end

command 2
# Save the rsp for now
set $savestack = $rsp
#c
end

command 3
#set $rsp = $savestack
#c
end

run
#q

#b *(((int($rip) & 0xfffffff00000) + 0x00007fff00000000 + 0x16670))
#b *(((int($rip) & 0xfffffff00000) + 0x00007fff00000000 + 0x16020))
#b *(((int($rip) & 0xfffffff00000) + 0x00007fff00000000 + 0x16e00))
#b *(((int($rip) & 0xfffffff00000) + 0x00007fff00000000 + 0x10b90))