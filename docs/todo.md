## TODO: 

- Stream output to WS
- Add execution timeouts for each request/operation
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

## Reading list

- runc internals
- containerd architecture
- gVisor syscall interception
- Firecracker microVMs
- Linux seccomp BPF filters
- Linux cgroups v2