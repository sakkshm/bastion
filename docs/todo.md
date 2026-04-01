# TODO: 

- add job timeout
- add job cancellation
- currently cancel works at api level but no actual change at conatiner level
- try using cgroups to kill/bunch jobs together
- add a container pool per session
- instead of persistent conatiners, think of conatiners as persistent compute pools and state is stored in volumes/mounts
- Look into daytona and open-terminal
- Stream output to WS
- Add execution timeouts for each request/operation/job

- learn about kernel level sandboxing, LXC componenets, cgroups, namespaces, seccomp, etc.
- Make an eBPF based tool for Bastion to help with sandboxing, conatiner sec/isolation/monitoring, enforcing network policy, loadbalancing, bandwidth management, sec and ops analysis. 
- Look into Linux seccomp BPF filters
- think about serverless, elastic infra, k8s compatible
- Read [Cisco-Cilium PPT](https://www.ciscolive.com/c/dam/r/ciscolive/global-event/docs/2025/pdf/DEVNET-2927.pdf) to get product inspo

- consider using cgroups or pid namespaces for job isolation and preventing escape. consider setsid, disown, double fork.
- look into /proc cleanup for zombie/orphan processes.
- Add dashboard type endpoint for admin to montior system
- Add a latency check/benchmark system
- Improve cmd exec latency by: use a persistent shell using stdin instead of calling exec always (one shell per worker), batching jobs
- Make sure only valid Session.status transition are allowed, (ex: deleted -> started not allowed)
- Make a container GC - Remove conatiners after TTL
- GC Remove conatiners once marked deleted
- Think about GC architechture
- Container creation is extremely slow, maybe look into making a conatiner pool and then assigning one on session creation
- Docker Containers can change state independetly and make Session.status stale, make some way to chage Session.status when Docker changes state
- Refactor code to make session data and docker container data always consistent, and make actions atomic
- Add comments/docs explaining handlers and flows
- Write a writeup on how this project uses docker and container isolation

- Look into Namespaces, Cgroups (CPU/Memory/PID limits), Seccomp syscall filtering, Linux capability dropping, Read-only root filesystem with writable sandbox mount, Network isolation, No-new-privileges, User namespace remapping (rootless containers), Filesystem mount restrictions, Device access restrictions, Execution timeouts

# Reading list

- runc internals
- containerd architecture
- gVisor syscall interception
- Firecracker microVMs
- Linux seccomp BPF filters
- Linux cgroups v2