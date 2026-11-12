# Disk Metrics Service

The `disk-metrics` service emits one snapshot-first payload per polling cycle.

The contract intentionally keeps three separate sections:

- `devices`: discovered local block-device-like entries such as disks,
  partitions, loop devices, device-mapper devices, and zvols
- `filesystems`: summaries of filesystem types observed from current mounts
- `mounts`: currently visible mounts, including network and memory-backed
  mounts that have no local block device

Those sections are related but not interchangeable:

- a device can exist without any mount
- a mount can exist without any local block device
- filesystem summaries are derived from observed mounts, not from installed
  kernel filesystem drivers
