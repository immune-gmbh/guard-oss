TITLE   cpu.asm: Assembly code for the x64 resources

.CODE chipsec_code$__a

PUBLIC _rdmsr

;------------------------------------------------------------------------------
;  void _rdmsr(
;    unsigned int msr_num, // rcx
;    unsigned int* msr_lo, // rdx
;    unsigned int* msr_hi  // r8
;    )
;------------------------------------------------------------------------------
_rdmsr PROC FRAME
    push r10
    .ALLOCSTACK 8
    push r11
    .ALLOCSTACK 8
    push rax
    .ALLOCSTACK 8
    push rdx
    .ALLOCSTACK 8
    .endprolog

    mov r10, rdx ; msr_lo
    mov r11, r8  ; msr_hi

    ; rcx has msr_num
    rdmsr

    ; Write MSR results in edx:eax
    mov dword ptr [r10], eax
    mov dword ptr [r11], edx

    pop rdx
    pop rax
    pop r11
    pop r10

    ret
_rdmsr ENDP

END
