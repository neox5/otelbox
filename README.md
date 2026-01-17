# obsbox

Telemetry signal generator for testing observability components. obsbox generates coherent metric signals derived from the same underlying data sources, enabling verification of monitoring pipelines, collectors, and transformation logic without requiring live systems.

Built on [simv](https://github.com/neox5/simv) for guaranteed mathematical consistency between related metrics.

## Quick Start

Run with Podman:

```bash
podman run -p 9090:9090 ghcr.io/neox5/obsbox:latest
```

Verify it's working:

```bash
curl http://localhost:9090/metrics
```

## Installation

### Container (Recommended)

```bash
podman pull ghcr.io/neox5/obsbox:latest
```

### Binary Releases

Download the latest release for your platform from [GitHub Releases](https://github.com/neox5/obsbox/releases):

```bash
# Linux (amd64)
wget https://github.com/neox5/obsbox/releases/download/v0.1.0/obsbox-linux-amd64
chmod +x obsbox-linux-amd64
./obsbox-linux-amd64 --version

# macOS (arm64)
wget https://github.com/neox5/obsbox/releases/download/v0.1.0/obsbox-darwin-arm64
chmod +x obsbox-darwin-arm64
./obsbox-darwin-arm64 --version
```

Verify checksums:

```bash
wget https://github.com/neox5/obsbox/releases/download/v0.1.0/obsbox-linux-amd64.sha256
sha256sum -c obsbox-linux-amd64.sha256
```

### Build from Source

Requires Go 1.25+:

```bash
go install github.com/neox5/obsbox/cmd/obsbox@latest
```

## Basic Usage

### Running with Default Configuration

The default configuration generates IBM MQ queue metrics:

**Podman:**

```bash
podman run -p 9090:9090 ghcr.io/neox5/obsbox:latest
```

**Binary:**

```bash
./obsbox
```

Access metrics at `http://localhost:9090/metrics`

### Running with Custom Configuration

Mount your configuration file:

**Podman:**

```bash
podman run -p 9090:9090 \
  -v $(pwd)/custom-config.yaml:/config/config.yaml:ro \
  ghcr.io/neox5/obsbox:latest \
  -config /config/config.yaml
```

**Binary:**

```bash
./obsbox -config custom-config.yaml
```

### CLI Flags

```
-config <path>    Path to configuration file (default: config.yaml)
--version         Print version and exit
```

## Configuration Guide

### Simple Example

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

### Configuration Sections

**simulation:** Defines value generation using [simv](https://github.com/neox5/simv)

- `clocks`: Time sources driving value updates
- `sources`: Raw value generators (random, constant, etc.)
- `values`: Derived values with transforms and views

**metrics:** Maps simulation values to exposed metrics

- `name`: Metric name (Prometheus or OTEL format)
- `type`: counter or gauge
- `value`: Reference to simulation value
- `attributes`: Labels/attributes for the metric

**export:** Configures metric exposition

- `prometheus`: Pull-based HTTP endpoint
- `otel`: Push-based OTLP export (future)

**settings:** Application settings

- `internal_metrics`: obsbox self-monitoring metrics

### Complete Example (IBM MQ Scenario)

Full-featured configuration showing all current capabilities:

```yaml
simulation:
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
      clone: v_mqput_counter
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
    enabled: false
    endpoint: localhost:4318
    interval: 10s
    resource:
      service.name: obsbox
      service.version: 0.1.0
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

- Multiple metrics from the same source (coherence guarantee)
- Counter and gauge representations of the same data
- Protocol-specific naming (Prometheus vs OTEL)
- Value derivation (cloning with reset behavior)
- Metric attributes/labels
- Both Prometheus and OTEL export configuration

### Common Patterns

**Adding Multiple Metrics:**
Reference the same value from multiple metric definitions to create coherent counter/gauge pairs.

**Changing Update Intervals:**
Modify clock interval in `simulation.clocks` section to control value generation rate.

**Multiple Independent Clocks:**
Define multiple named clocks with different intervals for independent metric groups.

## Configuration Reference

### Simulation Domain

Built on [simv](https://github.com/neox5/simv) library for value generation and transformation.

#### Clocks

Time sources driving value updates.

```yaml
clocks:
  <name>:
    type: periodic
    interval: <duration> # e.g., 1s, 500ms, 2m
```

**Types:**

- `periodic`: Fixed interval updates

#### Sources

Raw value generators attached to clocks.

```yaml
sources:
  <name>:
    type: random_int
    clock: <clock_name>
    min: <int>
    max: <int>
```

**Types:**

- `random_int`: Random integer in range [min, max]

#### Values

Derived values with transforms and views.

```yaml
values:
  <name>:
    source: <source_name>           # Create from source
    transforms: [<transform>...]    # Apply transforms

  <name>:
    clone: <value_name>             # Derive from existing value
    transforms: [<transform>...]    # Optional: extend transforms
    reset: on_read                  # Optional: reset behavior
```

**Transforms:**

- `accumulate`: Running sum (counter semantics)

**Reset Behavior:**

- `on_read`: Reset to zero after each read (gauge window semantics)
- `{type: on_read, value: <int>}`: Reset to specific value

**Coherence Guarantee:**
Multiple values derived from the same source maintain mathematical consistency. Metrics referencing these values produce coherent signals.

### Metrics Definition

Maps simulation values to exposed metrics.

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

- `counter`: Monotonically increasing value
- `gauge`: Value that can increase or decrease

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

Pull-based HTTP endpoint.

```yaml
export:
  prometheus:
    enabled: true | false
    port: <port> # Default: 9090
    path: <path> # Default: /metrics
```

#### OTEL (Future)

Push-based OTLP export.

```yaml
export:
  otel:
    enabled: true | false
    endpoint: <host:port>
    interval: <duration> # Push interval
    resource: # Resource attributes
      <key>: <value>
    headers: # Optional HTTP headers
      <key>: <value>
```

**Constraints:**

- Only one exporter can be enabled at a time
- At least one exporter must be enabled

### Settings

#### Internal Metrics

obsbox self-monitoring metrics.

```yaml
settings:
  internal_metrics:
    enabled: true | false
    format: native | underscore | dot
```

**Format:**

- `native`: Each exporter's native convention (underscore for Prometheus, dot for OTEL)
- `underscore`: Force underscore-separated names
- `dot`: Force dot-separated names

**Default:** `enabled: false`, `format: native`

## License

MIT License - see [LICENSE](LICENSE) file for details.
