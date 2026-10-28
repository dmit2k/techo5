// Independent native ASP2 IPC fixture. No vendor code, hardware, or real /data.
#define _GNU_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>
#ifndef CHILD_MARKER
#error CHILD_MARKER must distinguish original and decoy ELF
#endif
extern char **environ;
static uint16_t u16(const unsigned char *p) { return p[0] | (uint16_t)p[1]<<8; }
static uint32_t u32(const unsigned char *p) { return u16(p) | (uint32_t)u16(p+2)<<16; }
static uint64_t u64(const unsigned char *p) { return u32(p) | (uint64_t)u32(p+4)<<32; }
static void put16(unsigned char *p, uint16_t n) { p[0]=n; p[1]=n>>8; }
static void put32(unsigned char *p, uint32_t n) { put16(p,n); put16(p+2,n>>16); }
static void put64(unsigned char *p, uint64_t n) { put32(p,n); put32(p+4,n>>32); }
static int exact(int fd, void *buf, size_t len, int writing) {
    unsigned char *p=buf;
    while (len) {
        ssize_t n=writing ? write(fd,p,len) : read(fd,p,len);
        if (n<0 && errno==EINTR) continue;
        if (n<=0) return -1;
        p+=n; len-=n;
    }
    return 0;
}
static void stall(void) { for (;;) pause(); }
static uint32_t crc32(const unsigned char *p, size_t n) {
    uint32_t c=~0u;
    while (n--) { c^=*p++; for (int i=0;i<8;i++) c=(c>>1)^(0xedb88320u & (0u-(c&1))); }
    return ~c;
}
static int stream(const unsigned char *s, int mic, unsigned bits) {
    unsigned char want[16]={0};
    put16(want,mic?4:2); put16(want+2,1920); put32(want+4,96000);
    want[8]=mic?2:1; want[9]=mic?bits:15; want[10]=mic && bits==31; want[11]=mic?24:16;
    return !memcmp(s,want,16);
}
int main(void) {
    struct stat a,b,exe;
    if (fstat(3,&a) || fstat(4,&b) || !S_ISREG(a.st_mode) || !S_ISDIR(b.st_mode) ||
        stat("/proc/self/exe",&exe) || exe.st_dev!=a.st_dev || exe.st_ino!=a.st_ino) return 11;
    // Secure host must not inherit PATH, loader overrides, or transport ENV.
    int envs=0;
    for (char **p=environ; *p; p++,envs++)
        if (strcmp(*p,"LANG=C") && strcmp(*p,"LC_ALL=C")) return 12;
    if (envs!=2) return 13;
    int dir=openat(4,"etc",O_RDONLY|O_DIRECTORY|O_CLOEXEC|O_NOFOLLOW);
    if (dir<0) return 14;
    int fd=openat(dir,"profile.conf",O_RDONLY|O_CLOEXEC|O_NOFOLLOW|O_NONBLOCK);
    close(dir);
    if (fd<0 || fstat(fd,&a) || !S_ISREG(a.st_mode)) return 15;
    char profile[128]={0}, mode[32], proto[16]; int asset;
    ssize_t n=read(fd,profile,sizeof(profile)-1); close(fd);
    if (n<=0 || sscanf(profile,"%31s %15s %d",mode,proto,&asset)!=3) return 16;
    if (!strcmp(mode,"hold-start")) {
        // Bounded synthetic descendant keeps both pipes open after parent exit.
        pid_t pid=fork(); if (pid<0) return 17;
        if (!pid) { sleep(2); _exit(0); }
        return 0;
    }
    if (!strcmp(mode,"start-hang")) stall();
    if (exact(1,!strcmp(mode,"start-bad") ? "NOPE" : "ASP2",4,1)) return 18;
    if (!strcmp(mode,"start-bad")) stall();
    unsigned char cfg[8192]={0}; unsigned bits=0;
    if (exact(0,cfg,8,0) || memcmp(cfg,"CFG2",4) || u16(cfg+4)!=2) return 20;
    if (strcmp(proto,"cfg2")) return 32;
    {
        if (u16(cfg+6)!=16 || exact(0,cfg+8,8,0)) return 23;
        uint32_t total=u32(cfg+8); if (total<64 || total>8192 || u32(cfg+12)!=7) return 24;
        if (exact(0,cfg+16,total-16,0)) return 25;
        unsigned pos=16, prev=0, cores=0;
        while (pos<total) {
            if (total-pos<8) return 26;
            unsigned tag=u16(cfg+pos), flags=u16(cfg+pos+2), len=u32(cfg+pos+4); pos+=8;
            if (!tag || tag<=prev || len>total-pos || flags>1) return 27;
            if (tag<=2) {
                if (flags!=1 || len!=16) return 28;
                if (tag==1) bits=cfg[pos+9];
                if (!stream(cfg+pos,tag==1,bits)) return 29;
                cores++;
            } else if (flags) return 30;
            prev=tag;pos+=len;
        }
        if (cores!=2 || (bits!=23 && bits!=31)) return 31;
    }
    if (!strcmp(mode,"config-hang")) stall();
    unsigned char ack[12]={0}; memcpy(ack,"ACK2",4); ack[4]=1;
    if (!strcmp(mode,"ack-bad")) ack[0]='X';
    if (!strcmp(mode,"reject")) { ack[4]=2;ack[5]=1;put16(ack+6,522); }
    if (exact(1,ack,12,1)) return 33;
    if (!strcmp(mode,"closed-reader")) { close(0);stall(); }
    for (uint64_t seq=0;;seq++) {
        unsigned char h[20], pcm[38400];
        if (exact(0,h,20,0)) return 0;
        if (memcmp(h,"INP2",4) || u64(h+4)!=seq || u32(h+12)!=30720 || u32(h+16)!=7680) return 34;
        if (exact(0,pcm,sizeof(pcm),0)) return 35;
        if (!strcmp(mode,"response-hang")) stall();
        unsigned char reply[32+640]={0}; memcpy(reply,"RSP2",4);reply[4]=1;
        put64(reply+12,seq);put64(reply+20,seq*320);put32(reply+28,640);
        if (!strcmp(mode,"sequence")) put64(reply+12,seq+1);
        if (!strcmp(mode,"origin")) put64(reply+20,seq*320+320);
        if (!strcmp(mode,"fatal") || (!strcmp(mode,"runtime-fatal") && seq==4)) {
            reply[5]=2;put16(reply+6,503);
        }
        uint32_t seed=crc32(pcm,sizeof(pcm))+bits+(uint32_t)seq*997+CHILD_MARKER+asset;
        for (int i=0;i<320;i++) put16(reply+32+i*2,seed+i*37);
        if (exact(1,reply,sizeof(reply),1)) return 36;
    }
}
