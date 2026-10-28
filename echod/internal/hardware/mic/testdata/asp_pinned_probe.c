// Synthetic offline probe. No ASP IPC, vendor code, device paths or privileges.
#define _GNU_SOURCE
#include <dirent.h>
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

#ifndef PROBE_NONCE
#error PROBE_NONCE must independently identify the compiled executable
#endif

// Validate the complete relative path before opening anything. Each directory
// and final component is opened against the preceding pinned FD, never AT_FDCWD.
static int asset_open(const char *path) {
    char copy[4096];
    size_t len = strlen(path);
    if (!len || len >= sizeof(copy) || path[0] == '/' || path[len-1] == '/')
        return -1;
    memcpy(copy, path, len + 1);
    for (char *p = copy, *end; ; p = end + 1) {
        end = strchr(p, '/');
        if (end) *end = 0;
        if (!*p || !strcmp(p, ".") || !strcmp(p, "..")) return -1;
        if (!end) break;
        *end = '/';
    }
    int at = 4, owned = 0;
    for (char *p = copy, *end; ; p = end + 1) {
        end = strchr(p, '/');
        if (end) *end = 0;
        int flags = O_RDONLY | O_CLOEXEC | O_NOFOLLOW | O_NONBLOCK;
        if (end) flags |= O_DIRECTORY;
        int fd = openat(at, p, flags);
        if (owned) close(at);
        if (fd < 0) return -1;
        if (!end) {
            struct stat st;
            if (fstat(fd, &st) || !S_ISREG(st.st_mode)) {
                close(fd);
                return -1;
            }
            return fd;
        }
        at = fd;
        owned = 1;
    }
}

static int identity(int fd, const char *dev, const char *ino, struct stat *st) {
    return fstat(fd, st) == 0 &&
        (unsigned long long)st->st_dev == strtoull(dev, NULL, 10) &&
        (unsigned long long)st->st_ino == strtoull(ino, NULL, 10);
}

int main(int argc, char **argv) {
    if (argc != 7) return 10;
    struct stat bin, root, exe;
    if (!identity(3, argv[1], argv[2], &bin) || !S_ISREG(bin.st_mode)) return 11;
    if (!identity(4, argv[3], argv[4], &root) || !S_ISDIR(root.st_mode)) return 12;
    if (stat("/proc/self/exe", &exe) ||
        exe.st_dev != bin.st_dev || exe.st_ino != bin.st_ino) return 13;
    // Enumerate all inherited descriptors, excluding this enumeration's FD.
    DIR *dir = opendir("/proc/self/fd");
    if (!dir) return 14;
    struct dirent *ent;
    while ((ent = readdir(dir))) {
        char *end;
        long fd = strtol(ent->d_name, &end, 10);
        if (*end == 0 && fd >= 5 && fd != dirfd(dir)) {
            closedir(dir);
            return 15;
        }
    }
    closedir(dir);
    if (fcntl(3, F_GETFD) & FD_CLOEXEC || fcntl(4, F_GETFD) & FD_CLOEXEC)
        return 16;
    int fd = asset_open(argv[5]);
    if (fd < 0) return 30;
    char asset[128];
    ssize_t n = read(fd, asset, sizeof(asset)-1);
    close(fd);
    if (n < 0 || n == (ssize_t)sizeof(asset)-1) return 31;
    asset[n] = 0;
    printf("nonce=%s exe=%llu:%llu fd3=%llu:%llu fd4=%llu:%llu asset=%s\n",
        PROBE_NONCE, (unsigned long long)exe.st_dev, (unsigned long long)exe.st_ino,
        (unsigned long long)bin.st_dev, (unsigned long long)bin.st_ino,
        (unsigned long long)root.st_dev, (unsigned long long)root.st_ino, asset);
    fflush(stdout);
    if (!strcmp(argv[6], "sleep")) {
        for (;;) pause();
    }
    return 0;
}
