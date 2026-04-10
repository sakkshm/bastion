# TODO: 

~~- add job timeout~~
~~- Stream output to WS~~
- WS stream dies unexpectedly
- Add support for terminal resize
- Add support for envs
- Refactor handlers, they should only contain request related logic not actual business logic, that should be in utils, make abstraction better

- Add dashboard type endpoint for admin to montior system
- Add a latency check/benchmark system
- Make sure only valid Session.status transition are allowed, (ex: deleted -> started not allowed)

- use postgres to store state
- Make sure state is synced with system reality eg: if conatiner/fs is deleted, the session should also die

- Look into options for cmd exec
```
type ExecOptions struct {
    User         string   // User that will run the command
    Privileged   bool     // Is the container in privileged mode
    Tty          bool     // Attach standard streams to a tty.
    ConsoleSize  *[2]uint `json:",omitempty"` // Initial console size [height, width]
    AttachStdin  bool     // Attach the standard input, makes possible user interaction
    AttachStderr bool     // Attach the standard error
    AttachStdout bool     // Attach the standard output
    DetachKeys   string   // Escape keys for detach
    Env          []string // Environment variables
    WorkingDir   string   // Working directory
    Cmd          []string // Execution commands and args
}
```
- Look into hostconfig.pidmode, ipcmode, utsmode, cgroupmode
- Look into improving Hostconfig, conatinerconfig, network config

~~- Make a container GC - Remove conatiners after TTL~~
~~- GC Remove conatiners once marked deleted~~

- add a container pool per session
- instead of persistent conatiners, think of conatiners as persistent compute pools and state is stored in volumes/mounts
- Container creation is extremely slow, maybe look into making a conatiner pool and then assigning one on session creation
- Docker Containers can change state independetly and make Session.status stale, make some way to chage Session.status when Docker changes state
- Refactor code to make session data and docker container data always consistent, and make actions atomic

- Add comments/docs explaining handlers and flows
- Write a writeup on how this project uses docker and container isolation



# Reading list

- runc internals
- containerd architecture
- gVisor syscall interception
- Firecracker microVMs
- Linux seccomp BPF filters
- Linux cgroups v2
- Look into Namespaces, Cgroups (CPU/Memory/PID limits), Seccomp syscall filtering, Linux capability dropping, Read-only root filesystem with writable sandbox mount, Network isolation, No-new-privileges, User namespace remapping (rootless containers), Filesystem mount restrictions, Device access restrictions, Execution timeouts
- learn about kernel level sandboxing, LXC componenets, cgroups, namespaces, seccomp, etc.
- Make an eBPF based tool for Bastion to help with sandboxing, conatiner sec/isolation/monitoring, enforcing network policy, loadbalancing, bandwidth management, sec and ops analysis. 
- Look into Linux seccomp BPF filters
- think about serverless, elastic infra, k8s compatible
- Read [Cisco-Cilium PPT](https://www.ciscolive.com/c/dam/r/ciscolive/global-event/docs/2025/pdf/DEVNET-2927.pdf) to get product inspo
- consider using cgroups or pid namespaces for job isolation and preventing escape. consider setsid, disown, double fork.
- look into /proc cleanup for zombie/orphan processes.