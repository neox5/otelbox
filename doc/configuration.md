# Configuration Guide

Complete guide to configuring obsbox for telemetry signal generation.

## Configuration File Structure

obsbox uses YAML configuration files with four main sections:

- **simulation** - Defines value generation using [simv](https://github.com/neox5/simv)
- **metrics** - Maps simulation values to exposed metrics
- **export** - Configures metric exposition (Prometheus, OTEL)
- **settings** - Application settings

## Simple Example

Minimal configuration generating a single counter metric:

```yaml
simulation:
  clocks:
    tick:
      type: periodic
      interval: 1s

  sources:
    events:
      type: random_int
      clock: tick
      min: 0
      max: 10

  values:
    total_events:
      source: events
      transforms: [accumulate]

metrics:
  - name: app_events_total
    type: counter
    description: "Total events processed"
    value: total_events
    attributes:
      service: myapp

export:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
```

## Configuration Sections

### Simulation Domain

The simulation section defines how values are generated using the simv library.

#### Seed

Optional seed for reproducible simulations:

```yaml
simulation:
  seed: 12345
```

When omitted, a time-based seed is used (logged at startup for reproduction).

#### Clocks

Time sources driving value updates:

```yaml
clocks:
  <name>:
    type: periodic
    interval: <duration> # e.g., 1s, 500ms, 2m
```

Multiple clocks can be defined with different intervals for independent metric groups.

#### Sources

Raw value generators attached to clocks:

```yaml
sources:
  <name>:
    type: random_int
    clock: <clock_name>
    min: <int>
    max: <int>
```

Each source references a clock that drives its update schedule.

#### Values

Derived values with transforms and views:

```yaml
values:
  <name>:
    source: <source_name>
    transforms: [<transform>...]
    reset: on_read # Optional
```

**Available Transforms:**

- `accumulate` - Running sum (counter semantics)

**Reset Behavior:**

- `on_read` - Reset to zero after each read (gauge window semantics)
- `{type: on_read, value: <int>}` - Reset to specific value

**Coherence Guarantee:**
Multiple values derived from the same source maintain mathematical consistency. Metrics referencing these values produce coherent signals.

### Metrics Definition

Maps simulation values to exposed metrics:

```yaml
metrics:
  - name: <metric_name>              # Simple form (same name for both protocols)
    name:                             # Full form (protocol-specific names)
      prometheus: <prom_name>
      otel: <otel_name>
    type: counter | gauge
    description: <help_text>
    value: <value_name>               # Reference to simulation value
    attributes:                       # Labels (Prometheus) / Attributes (OTEL)
      <key>: <value>
```

**Metric Types:**

- `counter` - Monotonically increasing value
- `gauge` - Value that can increase or decrease

**Naming:**

- Simple form: Same name for Prometheus and OTEL
- Full form: Protocol-specific names (e.g., `app.events.total` vs `app_events_total`)

**Attributes:**

- Key-value pairs attached to metric
- Prometheus: labels
- OTEL: attributes
- Keys must match `[a-zA-Z_][a-zA-Z0-9_]*`
- Cannot start with `__`

### Export Configuration

#### Prometheus

Pull-based HTTP endpoint:

```yaml
export:
  prometheus:
    enabled: true | false
    port: <port> # Default: 9090
    path: <path> # Default: /metrics
```

#### OTEL

Push-based OTLP export:

```yaml
export:
  otel:
    enabled: true | false
    transport: grpc | http # Default: grpc
    host: <hostname> # Default: localhost
    port: <port> # Default: 4317 (grpc), 4318 (http)
    interval: <duration> # Push interval
    resource: # Resource attributes
      <key>: <value>
    headers: # Optional HTTP headers
      <key>: <value>
```

**Transport Types:**

- `grpc` - OTLP over gRPC (default, port 4317)
- `http` - OTLP over HTTP (port 4318)

**Constraints:**

- Only one exporter can be enabled at a time
- At least one exporter must be enabled

### Settings

#### Internal Metrics

obsbox self-monitoring metrics:

```yaml
settings:
  internal_metrics:
    enabled: true | false
    format: native | underscore | dot
```

**Format:**

- `native` - Each exporter's native convention (underscore for Prometheus, dot for OTEL)
- `underscore` - Force underscore-separated names
- `dot` - Force dot-separated names

**Default:** `enabled: false`, `format: native`

## Complete Example (IBM MQ Scenario)

Full-featured configuration showing all current capabilities:

```yaml
simulation:
  seed: 12345

  clocks:
    clk_mqput:
      type: periodic
      interval: 1s

  sources:
    src_mqput_events:
      type: random_int
      clock: clk_mqput
      min: 0
      max: 10

  values:
    v_mqput_counter:
      source: src_mqput_events
      transforms: [accumulate]

    v_mqput_count_reseted:
      source: src_mqput_events
      transforms: [accumulate]
      reset: on_read

metrics:
  - name:
      prometheus: ibmmq_queue_mqput_count_total
      otel: ibmmq.queue.mqput.count
    type: counter
    description: "Total MQ PUT operations since start"
    value: v_mqput_counter
    attributes:
      queue: QUEUE.1
      qmgr: QM1

  - name:
      prometheus: ibmmq_queue_mqput_count_ratio
      otel: ibmmq.queue.mqput.rate
    type: gauge
    description: "Recent MQ PUT operations in window"
    value: v_mqput_count_reseted
    attributes:
      queue: QUEUE.1
      qmgr: QM1

export:
  otel:
    enabled: true
    transport: grpc # grpc (default) or http
    host: localhost # default: localhost
    port: 4317 # default: 4317 (grpc), 4318 (http)
    interval: 10s
    resource:
      service.name: obsbox
      service.version: 0.2.0
      deployment.environment: development

  prometheus:
    enabled: true
    port: 9090
    path: /metrics

settings:
  internal_metrics:
    enabled: true
    format: native
```

This configuration demonstrates:

- Reproducible simulations with explicit seed
- Multiple metrics from the same source (coherence guarantee)
- Counter and gauge representations of the same data
- Protocol-specific naming (Prometheus vs OTEL)
- Value derivation with reset behavior
- Metric attributes/labels
- Both Prometheus and OTEL export configuration
- Transport selection for OTEL (gRPC or HTTP)

## Common Patterns

### Adding Multiple Metrics

Reference the same source from multiple value definitions to create coherent counter/gauge pairs:

```yaml
values:
  total_requests:
    source: request_events
    transforms: [accumulate]

  recent_requests:
    source: request_events
    transforms: [accumulate]
    reset: on_read

metrics:
  - name: http_requests_total
    type: counter
    value: total_requests

  - name: http_requests_current
    type: gauge
    value: recent_requests
```

### Changing Update Intervals

Modify clock interval in `simulation.clocks` section:

```yaml
clocks:
  fast_tick:
    type: periodic
    interval: 100ms # 10 updates per second

  slow_tick:
    type: periodic
    interval: 5s # 1 update every 5 seconds
```

### Multiple Independent Clocks

Define multiple named clocks with different intervals for independent metric groups:

```yaml
clocks:
  api_clock:
    type: periodic
    interval: 1s

  batch_clock:
    type: periodic
    interval: 30s

sources:
  api_requests:
    clock: api_clock
    # ...

  batch_jobs:
    clock: batch_clock
    # ...
```
