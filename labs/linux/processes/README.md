# Linux Processes

> A process is an executing instance of a program, together with the execution state and resources required to run it.

Understanding Linux processes is foundational to understanding containers, Kubernetes, CI runners, resource isolation, and production troubleshooting.

This lab explores what processes are, how they are created and terminated, how Linux tracks them, and how to investigate process-related failures.

---

## 1. Why Do Processes Exist?

A program is a passive set of instructions.

A process is what Linux creates when those instructions need to execute.

For example:

```text
/usr/bin/ls
```

is a program.

When a shell executes:

```bash
ls
```

Linux creates a process that executes the `ls` program.

Therefore:

```text
Program
  = instructions

Process
  = executing instance
    + execution state
    + resources
```

The same program can have multiple processes:

```text
             /usr/bin/python
                  │
        ┌─────────┼─────────┐
        ▼         ▼         ▼
    Process A  Process B  Process C
```

Each process has its own process identity and execution context.

---

# 2. What Does a Process Contain?

A process is more than just the program code.

Conceptually, it has:

```text
Process
├── PID
├── virtual address space
├── execution state
├── stack
├── heap
├── registers / CPU execution state
├── open file descriptors
├── environment variables
├── credentials
└── scheduling information
```

This becomes important later because many production failures are actually resource failures associated with processes:

* too many processes
* too much memory
* too many file descriptors
* CPU saturation
* excessive network connections

---

# 3. PID — Process ID

Every process has a **PID (Process ID)**.

Run:

```bash
echo $$
```

This prints the PID of the current shell.

You can inspect processes using:

```bash
ps
```

or:

```bash
ps -ef
```

Example:

```text
UID      PID   PPID  CMD
user    2056   927   bash
user    79537 2056   sleep 300
```

Here:

```text
PID  = identity of the process
PPID = identity of its immediate parent
```

---

# 4. PPID — Parent Process ID

Processes normally have a parent-child relationship.

For example:

```text
bash
 │
 └── ls
```

If:

```text
bash PID = 2056
ls PID   = 2100
```

then:

```text
ls PPID = 2056
```

PPID therefore tells us the **immediate parent** of a process.

It does not mean that the parent is necessarily PID 1.

For example:

```text
PID 1
 │
 └── bash
      │
      └── ls
```

Here:

```text
bash PPID = 1
ls   PPID = bash PID
```

---

# 5. PID 1

PID 1 is special.

It is the first userspace process in a Linux PID namespace and has special responsibilities, including acting as a process reaper in relevant circumstances.

On a typical Linux host, PID 1 is often:

```text
systemd
```

However, this is environment-dependent.

For example, inside a container, PID 1 may be a shell or another application.

During this experiment, the development container showed:

```text
PID 1
/bin/sh -c ...
```

rather than `systemd`.

This is an important reminder:

> Do not memorize "PID 1 = systemd."

Instead:

> **PID 1 is a special userspace process; what program occupies PID 1 depends on the environment.**

---

# 6. Process Hierarchy

Processes form relationships that can be represented as a tree.

Conceptually:

```text
PID 1
 │
 ├── service-A
 │    ├── worker
 │    └── worker
 │
 └── shell
      ├── command
      └── command
```

Inspect the process hierarchy with:

```bash
pstree -p
```

If `pstree` isn't installed, `ps` can still show the relationship:

```bash
ps -eo pid,ppid,stat,comm
```

The hierarchy becomes particularly useful when investigating:

* orphaned processes
* unexpected child processes
* process leaks
* CI runner cleanup problems
* service supervision

---

# 7. Process Creation

A traditional Unix process creation model uses:

```text
fork()
  ↓
child process
  ↓
exec()
  ↓
new program
```

### `fork()`

`fork()` creates a new process based on the calling process.

Conceptually:

```text
Before:

Parent


After:

Parent
  │
  └── Child
```

The child begins with inherited process state and then can follow a different execution path.

### `exec()`

`exec()` replaces the current process's program image with another program.

Importantly:

> **`exec()` does not create a new process.**

The PID remains associated with the same process while the program being executed changes.

Therefore:

```text
fork() → creates a process

exec() → replaces the process's program image
```

---

# 8. Why `fork()` + `exec()`?

Consider:

```bash
ls
```

The shell needs to execute `ls`.

Conceptually:

```text
bash
 │
 │ fork()
 ▼
child process
 │
 │ exec("ls")
 ▼
ls
```

The shell remains available to accept additional commands.

The child process becomes the `ls` program.

When `ls` exits, the parent shell can continue.

This separation between process creation and program execution is one of the foundational ideas of Unix process management.

---

# 9. Waiting for a Child

When a child process exits, its parent may need to collect its exit status.

This is done using mechanisms such as:

```text
wait()
waitpid()
```

Conceptually:

```text
Parent
  │
  └── Child
        │
        └── exit()
              │
              ▼
        exit status retained
              │
        parent wait()
              │
              ▼
           reaped
```

This leads directly to the concept of a zombie process.

---

# 10. Zombie Processes

A zombie is a process that has **already terminated**, but whose parent has not yet collected its termination status.

Conceptually:

```text
Child
  │
  └── exit()
        │
        ▼
     ZOMBIE
        │
        │ parent calls wait()
        ▼
     REAPED
```

A zombie:

* is no longer executing
* does not consume CPU like a running process
* remains represented in the process table until it is reaped

### Important misconception

A zombie is **not** a process that is still running in the background.

It has already exited.

---

# 11. Orphan Processes

An orphan is different.

An orphan is a process whose parent exits while the child is still running.

```text
Parent
  X  ← exits

Child
  │
  ▼
still running
```

The child is then reparented to an appropriate reaper process.

Conceptually:

```text
ZOMBIE

child dies
    ↓
parent still alive
    ↓
child waits to be reaped


ORPHAN

parent dies
    ↓
child still alive
    ↓
child gets reparented
```

### Key distinction

```text
Zombie = child is DEAD, parent has not reaped it

Orphan = child is ALIVE, parent has died
```

Do not confuse the two.

---

# 12. Process States

Inspect process state with:

```bash
ps -eo pid,ppid,stat,comm
```

Common states include:

| State | Meaning               |
| ----- | --------------------- |
| `R`   | Running or runnable   |
| `S`   | Interruptible sleep   |
| `D`   | Uninterruptible sleep |
| `T`   | Stopped or traced     |
| `Z`   | Zombie                |

### Why is `D` important?

A process in `D` state is generally blocked in an uninterruptible sleep, commonly while waiting for kernel-level I/O or another operation that cannot safely be interrupted at that point.

If a production process is:

```text
CPU = 0%
STATE = D
```

do not conclude that everything is healthy.

Investigate:

```text
What is it waiting for?
What resource is involved?
Is there an I/O problem?
Is storage unavailable or slow?
Is the process stuck in a kernel operation?
```

The state is a clue, not the root cause.

---

# 13. Signals

Processes need a mechanism for receiving asynchronous notifications.

Linux uses **signals**.

Examples include:

```text
SIGTERM
SIGKILL
SIGINT
SIGHUP
SIGSTOP
SIGCONT
```

The `kill` command is primarily a signal-sending utility.

For example:

```bash
kill -TERM <PID>
```

sends `SIGTERM` to the target process.

When no signal is specified:

```bash
kill <PID>
```

the default signal is normally `SIGTERM`.

---

# 14. SIGTERM vs SIGKILL

## SIGTERM

`SIGTERM` requests graceful termination.

An application can respond by performing appropriate shutdown work such as:

* finishing or safely stopping work
* closing resources
* stopping workers
* flushing appropriate state
* releasing resources

The important operational idea is:

```text
SIGTERM
   ↓
request graceful shutdown
   ↓
observe whether shutdown completes
```

## SIGKILL

```bash
kill -KILL <PID>
```

or:

```bash
kill -9 <PID>
```

requests immediate termination by the kernel.

`SIGKILL` cannot be caught or handled by the application.

Therefore:

```text
SIGTERM = graceful termination request

SIGKILL = forceful termination
```

### Operational rule

Do not automatically use:

```bash
kill -9
```

just because a process hasn't exited immediately.

First investigate why graceful termination failed.

Potential causes include:

* blocked I/O
* deadlock
* shutdown bugs
* stuck dependencies
* resource problems
* application bugs

Before using `SIGKILL`, capture useful evidence where practical.

Afterward, verify that the service actually recovered.

---

# 15. `/proc/<PID>` — Process Information

Linux exposes process information through `/proc`.

For a process with PID `79537`:

```bash
cat /proc/79537/status
```

Useful information includes:

```text
Name
State
Pid
PPid
Uid
Gid
Threads
```

You can inspect the command line:

```bash
tr '\0' ' ' < /proc/79537/cmdline
echo
```

This provides another view of what the process was started with.

`/proc` becomes particularly important later when investigating:

* memory
* file descriptors
* CPU
* networking
* namespaces
* process state

---

# 16. Hands-on Experiment

A simple process lifecycle experiment:

```bash
sleep 300 &
```

Find the process:

```bash
pgrep -n sleep
```

Inspect it:

```bash
ps -p <PID> -o pid,ppid,stat,etime,cmd
```

Inspect its process metadata:

```bash
cat /proc/<PID>/status
```

Inspect its command line:

```bash
tr '\0' ' ' < /proc/<PID>/cmdline
echo
```

Terminate it gracefully:

```bash
kill -TERM <PID>
```

Verify:

```bash
ps -p <PID>
```

### Observed lifecycle

```text
shell
  │
  └── sleep
        │
        ├── PID assigned
        ├── process inspected
        └── SIGTERM
              ↓
           process exits
```

---

# 17. Production Troubleshooting

A process symptom is not automatically a root cause.

A useful investigation sequence is:

```text
Symptom
   ↓
Identify process
   ↓
PID / PPID
   ↓
Process state
   ↓
CPU / memory
   ↓
Threads
   ↓
Open files / sockets
   ↓
I/O / system calls
   ↓
Logs
   ↓
Dependencies
   ↓
Root cause
   ↓
Mitigation
   ↓
Prevention
```

The most important habit is:

> **Do not convert an observation directly into a conclusion.**

For example:

### Bad reasoning

```text
50 processes remain
    ↓
They must be zombies
```

### Better reasoning

```text
50 processes remain
    ↓
Inspect their states
    ↓
Identify parent/child relationships
    ↓
Determine whether they are:
    running / sleeping / blocked / orphaned / zombies
    ↓
Investigate the actual failure
```

The same principle applies to CPU.

### Bad reasoning

```text
CPU = 100%
    ↓
Service is broken
```

### Better reasoning

```text
CPU = 100%
    ↓
What capacity does 100% represent?
    ↓
Which process/thread is consuming it?
    ↓
Is it sustained?
    ↓
Are latency/errors/queues increasing?
    ↓
Is there actual user impact?
```

This evidence-first approach is a core production engineering skill.

---

# 18. Connection to CI Runner Infrastructure

CI runners execute arbitrary workloads.

A simplified model is:

```text
CI Job
   │
   ▼
Runner
   │
   ├── git
   ├── compiler
   ├── tests
   ├── packaging
   └── artifact signing
```

All of these ultimately involve processes.

This creates several problems if arbitrary build code runs directly on a shared host:

### Resource exhaustion

A build could consume:

```text
CPU
Memory
PIDs
Disk
File descriptors
Network connections
```

### Filesystem isolation

Build A should not be able to modify Build B's workspace.

### Credential isolation

A build should not automatically gain access to credentials belonging to the runner or another workload.

### Process isolation

A build should not be able to interfere with unrelated workloads.

### Lifecycle cleanup

When a build ends, its processes and resources should not remain behind and contaminate the next build.

This motivates stronger isolation mechanisms such as:

```text
containers
VMs
namespaces
cgroups
ephemeral runners
```

These topics will be studied later.

---

# 19. Common Interview Traps

### Trap 1

**Does `exec()` create a new process?**

No.

```text
fork() → process creation
exec() → program-image replacement
```

### Trap 2

**Does a zombie consume CPU?**

No.

It has already terminated.

### Trap 3

**Is every leftover process a zombie?**

No.

It may be running, sleeping, blocked, orphaned, or a zombie.

### Trap 4

**Is PID 1 the direct parent of every process?**

No.

PPID identifies the immediate parent.

### Trap 5

**Does 100% CPU automatically mean an incident?**

No.

Determine the measurement context and whether there is actual service impact.

### Trap 6

**Should `kill -9` be the default way to stop a stuck process?**

No.

Investigate first and use forceful termination when operationally justified.

---

# 20. Mental Model

The most useful mental model from this lab is:

```text
              PROGRAM
                 │
                 │ execute
                 ▼
              PROCESS
                 │
       ┌─────────┼─────────┐
       │         │         │
      PID      PPID     RESOURCES
       │                   │
       │             ┌─────┼─────┐
       │             │     │     │
       │            CPU  memory  FDs
       │
       ▼
   PROCESS TREE
       │
       ├── child
       ├── child
       └── child
             │
             ▼
        process states
             │
       ┌─────┼─────┐
       │     │     │
       R     S     D
                   │
                   └── investigate what it is waiting for

       exit
        │
        ├── reaped → gone
        │
        └── not reaped → zombie
```

---

# 21. Key Lessons

1. A **program** is passive instructions; a **process** is an executing instance with state and resources.
2. Every process has a PID; PPID identifies its immediate parent.
3. `fork()` creates a child process.
4. `exec()` replaces the current process's program image.
5. `wait()` / `waitpid()` allow a parent to collect child termination status.
6. A zombie has already exited but has not been reaped.
7. An orphan is still running after its parent exits and is subsequently reparented.
8. Process state is an important troubleshooting signal.
9. `kill` sends signals; it does not inherently mean "destroy."
10. `SIGTERM` requests graceful termination; `SIGKILL` forcefully terminates.
11. `/proc/<PID>` exposes valuable process information.
12. High resource utilization is a **symptom**, not automatically a root cause.
13. Production troubleshooting should follow **observation → evidence → hypothesis → root cause → mitigation → prevention**.
14. CI runners need process, resource, filesystem, credential, and lifecycle isolation when executing untrusted workloads.

---

# 22. What Comes Next?

The next layer is **File Descriptors**.

Processes don't operate in isolation. They interact with the outside world through resources such as files, pipes, terminals, and network sockets.

That leads to:

```text
Process
   │
   ▼
File Descriptors
   │
   ├── stdin
   ├── stdout
   ├── stderr
   ├── files
   ├── pipes
   └── sockets
        │
        ▼
FD exhaustion
        │
        ▼
Production failures
        │
        ▼
CI runner reliability
```

The next lab will deliberately reproduce **file-descriptor exhaustion** and diagnose it using Linux tooling.

