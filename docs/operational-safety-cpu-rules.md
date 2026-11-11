# Operational Safety CPU frequency rules

The [four-rule Raspberry Pi 5 example](examples/operational-safety-rpi-cpu-rules.json) links three protective
stages and one normal record. It is **an example, not the installed default**. The packaged
`/etc/lite-nas/operational-safety-automation/rules.json` contains `{"rules":[]}`. Choose caps supported by the
actual host before copying the example into that file.

## Source events already defined by the monitors

Sensor Monitor publishes readings; Resources Monitor turns the following packaged sensor rules into system
alerts. Each alert ID starts with the shown prefix and has a generated suffix. The event is normalized when
that rule's sensor value leaves its band.

| Condition | Event ID prefix | Source | Category | Band |
| --- | --- | --- | --- | --- |
| 5 V warning | `snsvwrn` | `sensor-monitor` | `sensor.power.input` | 4.85 V ≤ value < 4.9 V |
| 5 V error | `snsverr` | `sensor-monitor` | `sensor.power.input` | 4.81 V < value < 4.85 V |
| 5 V critical | `snsvcrit` | `sensor-monitor` | `sensor.power.input` | value ≤ 4.81 V |
| CPU temperature warning | `snstwrn` | `sensor-monitor` | `sensor.temperature.cpu` | 65 °C ≤ value < 70 °C |
| CPU temperature error | `snsterr` | `sensor-monitor` | `sensor.temperature.cpu` | 70 °C ≤ value < 75 °C |
| CPU temperature critical | `snstcrit` | `sensor-monitor` | `sensor.temperature.cpu` | value ≥ 75 °C |

These are defined in
[`configs/etc/lite-nas/resources-monitor/rules/sensor-monitor.json`](../configs/etc/lite-nas/resources-monitor/rules/sensor-monitor.json).
A sensor that is absent from a snapshot cannot produce its alert. The voltage key is `EXT5V_V`; the
temperature key is `thermal:cpu-thermal`.

## Four-rule chain

The example uses a five-minute cooldown at each protective stage. Every protective rule has `active_khz`, the
cap to verify while its source alert is active, and `normal_khz`, the cap to verify after its source condition
clears and its cooldown finishes. The output ID prefixes are unique and must not match the source event
prefixes.

| Rule | Source alerts that can activate it | Required parent event | Active cap | Recovery cap | Owned event |
| --- | --- | --- | ---: | ---: | --- |
| `cpu_warning` | warning, error, or critical 5 V or CPU temperature | none | 1,800,000 kHz | 1,900,000 kHz | `opswarn` |
| `cpu_error` | error or critical 5 V or CPU temperature | active `opswarn` | 1,600,000 kHz | 1,800,000 kHz | `opserr` |
| `cpu_critical` | critical 5 V or CPU temperature | active `opserr` | 1,500,000 kHz | 1,600,000 kHz | `opscrit` |
| `cpu_normal` | normalized `opswarn` | none | — | verified 1,900,000 kHz | `opsnorm` |

The `cpu_normal` rule records the verified restoration. It publishes an informational event and immediately
requests its normalization; it does not set the cap again.

A sensor can jump directly into the critical band. The warning rule therefore accepts warning **and more
severe** source alerts, and the error rule accepts error **and critical** source alerts. The stages still
apply in order: 1800 → 1600 → 1500 MHz. Each successful cap change publishes its owned event, which returns
through the service's NATS subscription and allows the next stage to start. A child stage keeps its parent
event active while the child is active or cooling.

Recovery follows the reverse order. After all critical source alerts normalize, `cpu_critical` waits five
uninterrupted minutes, verifies 1600 MHz, and only then normalizes its `opscrit` event. `cpu_error` can start
its own five-minute cooldown only after its source alerts and the critical stage have cleared; it then
verifies 1800 MHz and normalizes `opserr`. `cpu_warning` similarly verifies the configured 1900 MHz cap before
normalizing `opswarn`. A newly active bad-value alert cancels that stage's cooldown; a fresh full cooldown
starts after the condition clears again. A failed cap write or read-back leaves the stage event active.

## Check the Pi before selecting caps

The caps are maximum CPU policy limits, in **kHz**. They do not force the CPU to run at that frequency. On the
target host, inspect every policy:

```sh
for policy in /sys/devices/system/cpu/cpufreq/policy*; do
    echo "$policy"
    for field in cpuinfo_min_freq cpuinfo_max_freq scaling_min_freq scaling_max_freq; do
        printf '%s=' "$field"
        cat "$policy/$field"
    done
done
```

Each proposed cap must be at least the policy's `cpuinfo_min_freq` and `scaling_min_freq`, and no higher than
`cpuinfo_max_freq`. The helper writes each policy's `scaling_max_freq` and reads it back; it rejects an
unsupported or rounded result. A first automatic change may only tighten the current cap. The reported Pi 5
lists 1500, 1600, 1800, and 1900 MHz in `scaling_available_frequencies`; successful writes and read-backs
still need a live test.

Raspberry Pi's published defaults vary: many Pi 4 revisions have a 1500 or 1800 MHz maximum, while the Pi 5
family lists a 1500 MHz default minimum and 2400 MHz default maximum. This example is tailored to the reported
Pi 5 and its operator-selected 1900 MHz ceiling for a system with two external SSDs. It is **not a portable
default**. Check the live policy files and choose supported values; changing the example does not overclock
the board. See the [Raspberry Pi frequency
defaults](https://www.raspberrypi.com/documentation/computers/config_txt.html#overclocking-options) and [Linux
CPUFreq policy attributes](https://docs.kernel.org/admin-guide/pm/cpufreq.html).

## Enroll and recover

The reported Pi 5 currently has `scaling_max_freq=1500000`. This example cannot start a protective descent
from that cap. Review why the cap is 1500 MHz before changing it. To set the example's 1900 MHz normal ceiling
manually with the standard Linux CLI, run:

```sh
sudo cpupower frequency-set --max 1900MHz
cpupower frequency-info
```

For a manual 2100 MHz ceiling, use `sudo cpupower frequency-set --max 2100MHz`. Use the same command with
`1500MHz`, `1600MHz`, or `1800MHz` to select one of the example's protective ceilings. The operational-safety
service applies configured rule caps itself; these commands are for manual host changes and checks. Install
the `cpupower` utility for your distribution if it is not present. The frequency must be supported by every
CPU policy.

Copy the reviewed example to `/etc/lite-nas/operational-safety-automation/rules.json` and restart
`lite-nas-operational-safety-automation.service` only when no automation task is active. Startup loads and
validates the rules but does not set a CPU cap or recover old tasks. Existing alerts from before the restart
are not replayed into a new task. The operator command also provides an explicit verified cap change after a
restart.

The example is tested with a virtual clock and simulated CPU policies. Before production enrollment, verify
the actual Pi policy limits and observe one warning, escalation, and full recovery on that host.
