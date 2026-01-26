# Configuration Reference

Detailed parameter reference for obsbox configuration files.

## File Format

YAML format with four top-level sections:

- `simulation`
- `metrics`
- `export`
- `settings`

## Simulation Domain

Built on [simv](https://github.com/neox5/simv) library for value generation and transformation.

### Seed

Optional seed configuration for reproducible simulations.

```yaml
simulation:
  seed: <uint64>
```

**Parameters:**

- `seed` (uint64, optional) - Master seed for random number generation
  - When omitted, uses time-based seed (logged at startup)
  - Same seed produces identical value sequences across runs
  - Example: `seed: 12345`

### Clocks

Time sources driving value updates.

```yaml
clocks:
  <name>:
    type: periodic
    interval: <duration>
```

**Parameters:**

- `type` (string, required) - Clock type
  - Values: `periodic`
- `interval` (duration, required) - Update interval
  - Format: Go duration string (e.g., `1s`, `500ms`, `2m`)
  - Examples: `100ms`, `1s`, `5s`, `1m`

### Sources

Raw value generators attached to clocks.

```yaml
sources:
  <name>:
    type: random_int
    clock: <clock_name>
    min: <int>
    max: <int>
```

**Parameters:**

- `type` (string, required) - Source type
  - Values: `random_int`
- `clock` (string, required) - Clock reference
  - Must reference existing clock name
- `min` (int, required for random_int) - Minimum value (inclusive)
- `max` (int, required for random_int) - Maximum value (inclusive)

### Values

Derived values with transforms and views.

```yaml
values:
  <name>:
    source: <source_name>
    transforms: [<transform>...]
    reset: <reset_config>
```

**Parameters:**

- `source` (string, required) - Source reference
  - Must reference existing source name
- `transforms` (array[transform], optional) - Transform pipeline
  - Applied in order
  - Available transforms: `accumulate`
- `reset` (reset_config, optional) - Reset behavior
  - See Reset Configuration below

**Transform Types:**

- `accumulate` - Running sum transform
  - Converts stream to monotonically increasing counter

**Reset Configuration:**

Short form:

```yaml
reset: on_read
```

Full form:

```yaml
reset:
  type: on_read
  value: <int> # Default: 0
```

- `type` (string) - Reset trigger
  - Values: `on_read`
- `value` (int, optional) - Reset target value
  - Default: 0

## Metrics Definition

Maps simulation values to exposed metrics.

```yaml
metrics:
  - name: <metric_name>
    name:
      prometheus: <prom_name>
      otel: <otel_name>
    type: <metric_type>
    description: <help_text>
    value: <value_name>
    attributes:
      <key>: <value>
```

**Parameters:**

- `name` (string or object, required) - Metric name
  - Short form: Single string (same for both protocols)
  - Full form: Object with `prometheus` and `otel` keys
- `type` (string, required) - Metric type
  - Values: `counter`, `gauge`
- `description` (string, required) - Metric help text
  - Alias: `help` (deprecated, use `description`)
- `value` (string, required) - Value reference
  - Must reference existing value name
- `attributes` (map[string]string, optional) - Metric labels/attributes
  - Alias: `labels` (deprecated, use `attributes`)
  - Keys must match `[a-zA-Z_][a-zA-Z0-9_]*`
  - Keys cannot start with `__`

**Metric Types:**

- `counter` - Monotonically increasing value
  - Used for cumulative metrics (total requests, bytes sent)
  - Value never decreases
- `gauge` - Value that can increase or decrease
  - Used for current state metrics (active connections, queue depth)
  - Value can go up or down

## Export Configuration

### Prometheus

Pull-based HTTP endpoint.

```yaml
export:
  prometheus:
    enabled: <bool>
    port: <int>
    path: <string>
```

**Parameters:**

- `enabled` (bool, required) - Enable Prometheus exporter
- `port` (int, optional) - HTTP port
  - Default: 9090
  - Range: 1-65535
- `path` (string, optional) - Metrics endpoint path
  - Default: `/metrics`

### OTEL

Push-based OTLP export.

```yaml
export:
  otel:
    enabled: <bool>
    transport: <transport_type>
    host: <string>
    port: <int>
    interval: <interval_config>
    resource:
      <key>: <value>
    headers:
      <key>: <value>
```

**Parameters:**

- `enabled` (bool, required) - Enable OTEL exporter
- `transport` (string, optional) - OTLP transport protocol
  - Values: `grpc`, `http`
  - Default: `grpc`
- `host` (string, optional) - OTLP endpoint host
  - Default: `localhost`
- `port` (int, optional) - OTLP endpoint port
  - Default: 4317 (grpc), 4318 (http)
- `interval` (interval_config, required) - Export intervals
  - See Interval Configuration below
- `resource` (map[string]string, optional) - Resource attributes
  - Default: `service.name: obsbox`, `service.version: dev`
- `headers` (map[string]string, optional) - Custom HTTP headers

**Interval Configuration:**

Short form (same read and push interval):

```yaml
interval: 10s
```

Full form (different intervals):

```yaml
interval:
  read: 1s
  push: 10s
```

- `read` (duration) - How often to read values
- `push` (duration) - How often to push to collector

**Default Resource Attributes:**

- `service.name` - Default: `obsbox`
- `service.version` - Default: `dev`

**Transport Types:**

- `grpc` - OTLP over gRPC
  - Default port: 4317
  - Standard OTEL collector gRPC endpoint
- `http` - OTLP over HTTP
  - Default port: 4318
  - Standard OTEL collector HTTP endpoint

### Export Constraints

- At least one exporter must be enabled
- Only one exporter can be enabled at a time (prevents read conflicts)

## Settings

Application-level settings.

```yaml
settings:
  internal_metrics:
    enabled: <bool>
    format: <naming_format>
```

**Parameters:**

- `internal_metrics.enabled` (bool, optional) - Enable obsbox self-monitoring
  - Default: `false`
- `internal_metrics.format` (string, optional) - Naming convention
  - Values: `native`, `underscore`, `dot`
  - Default: `native`

**Naming Formats:**

- `native` - Use each exporter's native convention
  - Prometheus: underscore-separated (e.g., `obsbox_metric_name`)
  - OTEL: dot-separated (e.g., `obsbox.metric.name`)
- `underscore` - Force underscore-separated names for all exporters
- `dot` - Force dot-separated names for all exporters

## Data Types

### Duration

Go duration string format:

- Valid units: `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`
- Examples: `100ms`, `1s`, `5s`, `30s`, `1m`, `2h`
- Can combine units: `1m30s`

### Integer

Signed 64-bit integer:

- Range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807
- No quotes in YAML

### String

UTF-8 text:

- Quote if contains special characters
- No length limit

### Boolean

Boolean value:

- Values: `true`, `false`
- Case-insensitive in YAML

### Map

Key-value pairs:

```yaml
key1: value1
key2: value2
```

### Array

Ordered list:

```yaml
- item1
- item2
```

Or inline:

```yaml
[item1, item2]
```
