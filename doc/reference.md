# Configuration Reference

Detailed parameter reference for obsbox configuration files.

## File Format

YAML format with five top-level sections:

- `templates`
- `instances`
- `metrics`
- `export`
- `settings`

## Templates

Reusable configuration definitions that support overrides when referenced.

### Templates.Clocks

Clock templates define reusable timing patterns.

```yaml
templates:
  clocks:
    <name>:
      type: <clock_type>
      interval: <duration>
```

**Parameters:**

- `type` (string, required) - Clock type
  - Values: `periodic`
- `interval` (duration, required) - Update interval
  - Format: Go duration string (e.g., `1s`, `500ms`, `2m`)

**Usage:**
Referenced in source templates or inline definitions:

```yaml
source:
  clock:
    template: <clock_template_name>
```

### Templates.Sources

Source templates define reusable data generators.

```yaml
templates:
  sources:
    <name>:
      type: <source_type>
      clock: <clock_reference>
      min: <int>
      max: <int>
```

**Parameters:**

- `type` (string, required) - Source type
  - Values: `random_int`
- `clock` (clock_reference, required) - Clock reference
  - See Clock Reference Types
- `min` (int, required for random_int) - Minimum value (inclusive)
- `max` (int, required for random_int) - Maximum value (inclusive)

**Usage:**
Referenced in value templates or inline definitions:

```yaml
value:
  source:
    template: <source_template_name>
    max: 200 # Optional override
```

### Templates.Values

Value templates define reusable transformation pipelines.

```yaml
templates:
  values:
    <name>:
      source: <source_reference>
      transforms: [<transform>...]
      reset: <reset_config>
```

**Parameters:**

- `source` (source_reference, required) - Source reference
  - See Source Reference Types
- `transforms` (array[transform], optional) - Transform pipeline
  - Applied in order
  - Available transforms: `accumulate`
- `reset` (reset_config, optional) - Reset behavior
  - See Reset Configuration

**Usage:**
Referenced in metrics or inline:

```yaml
metrics:
  - value:
      template: <value_template_name>
      reset: on_read # Optional override
```

### Templates.Metrics

Metric templates define reusable metric patterns (type, attributes).

```yaml
templates:
  metrics:
    <name>:
      type: <metric_type>
      value: <value_reference>
      attributes:
        <key>: <value>
```

**Parameters:**

- `type` (string, required) - Metric type
  - Values: `counter`, `gauge`
- `value` (value_reference, optional) - Value reference
  - Can be overridden when template is used
- `attributes` (map[string]string, optional) - Metric attributes
  - Can be overridden when template is used

**Usage:**
Not yet implemented in metrics section (reserved for future use).

## Instances

Named, concrete objects that can be shared across references without modification.

### Instances.Clocks

Clock instances define shared timing sources.

```yaml
instances:
  clocks:
    <name>:
      type: <clock_type>
      interval: <duration>
```

**Parameters:**
Same as Templates.Clocks

**Usage:**
Referenced by name in sources:

```yaml
source:
  clock:
    instance: <clock_instance_name>
```

**Behavior:**

- All references share the same clock instance
- Updates synchronized across all references
- No overrides allowed

### Instances.Sources

Source instances define shared data generators.

```yaml
instances:
  sources:
    <name>:
      type: <source_type>
      clock: <clock_reference>
      min: <int>
      max: <int>
```

**Parameters:**
Same as Templates.Sources

**Usage:**
Referenced by name in values:

```yaml
value:
  source:
    instance: <source_instance_name>
```

**Behavior:**

- All references share the same source instance
- Guarantees data coherence across metrics
- No overrides allowed

### Instances.Values

Value instances define shared transformation pipelines.

```yaml
instances:
  values:
    <name>:
      source: <source_reference>
      transforms: [<transform>...]
      reset: <reset_config>
```

**Parameters:**
Same as Templates.Values

**Usage:**
Referenced by name in metrics:

```yaml
metrics:
  - value:
      instance: <value_instance_name>
```

**Behavior:**

- All references share the same value instance
- No overrides allowed

## Reference Types

Configuration fields that reference other objects support three forms:

### Clock Reference

**Instance Reference:**

```yaml
clock:
  instance: <clock_instance_name>
```

**Template Reference:**

```yaml
clock:
  template: <clock_template_name>
  interval: <duration> # Optional override
```

**Inline Definition:**

```yaml
clock:
  type: periodic
  interval: <duration>
```

### Source Reference

**Instance Reference:**

```yaml
source:
  instance: <source_instance_name>
```

**Template Reference:**

```yaml
source:
  template: <source_template_name>
  clock: <clock_reference> # Optional override
  min: <int> # Optional override
  max: <int> # Optional override
```

**Inline Definition:**

```yaml
source:
  type: random_int
  clock: <clock_reference>
  min: <int>
  max: <int>
```

### Value Reference

**Instance Reference:**

```yaml
value:
  instance: <value_instance_name>
```

**Template Reference:**

```yaml
value:
  template: <value_template_name>
  source: <source_reference> # Optional override
  transforms: [<transform>...] # Optional override
  reset: <reset_config> # Optional override
```

**Inline Definition:**

```yaml
value:
  source: <source_reference>
  transforms: [<transform>...]
  reset: <reset_config>
```

## Metrics Definition

Maps values to exposed metrics.

```yaml
metrics:
  - name: <metric_name>
    name:
      prometheus: <prom_name>
      otel: <otel_name>
    type: <metric_type>
    description: <help_text>
    value: <value_reference>
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
- `value` (value_reference, required) - Value reference
  - See Value Reference Types
- `attributes` (map[string]string, optional) - Metric labels/attributes
  - Keys must match `[a-zA-Z_][a-zA-Z0-9_]*`
  - Keys cannot start with `__`

**Metric Types:**

- `counter` - Monotonically increasing value
  - Used for cumulative metrics (total requests, bytes sent)
  - Value never decreases
- `gauge` - Value that can increase or decrease
  - Used for current state metrics (active connections, queue depth)
  - Value can go up or down

## Transform Configuration

### Transform Types

**Accumulate Transform:**

Short form:

```yaml
transforms: [accumulate]
```

Full form:

```yaml
transforms:
  - type: accumulate
```

Converts stream to monotonically increasing counter (running sum).

## Reset Configuration

Defines when and how values reset.

**Short form (reset to zero):**

```yaml
reset: on_read
```

**Full form (reset to specific value):**

```yaml
reset:
  type: on_read
  value: <int>
```

**Parameters:**

- `type` (string, required) - Reset trigger
  - Values: `on_read`
- `value` (int, optional) - Reset target value
  - Default: 0

**Behavior:**

- `on_read` - Value resets after each read operation
  - Useful for gauge semantics (window-based metrics)
  - Counter maintains history, gauge shows recent activity

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
  seed: <uint64>
  internal_metrics:
    enabled: <bool>
    format: <naming_format>
```

**Parameters:**

- `seed` (uint64, optional) - Master seed for random number generation
  - When omitted, uses time-based seed (logged at startup)
  - Same seed produces identical value sequences across runs
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

## Reference Resolution

### Override Behavior

When templates are referenced, specific fields can be overridden:

```yaml
templates:
  sources:
    base:
      type: random_int
      clock:
        type: periodic
        interval: 1s
      min: 0
      max: 100

# Override max only
metrics:
  - value:
      source:
        template: base
        max: 200 # Override, other fields from template
```

**Override Rules:**

- Only specified fields are overridden
- Unspecified fields use template values
- Nested objects can be partially overridden
- Arrays are replaced entirely (not merged)

### Instance Restrictions

Instances cannot be overridden:

```yaml
instances:
  sources:
    shared:
      type: random_int
      clock:
        instance: main_tick
      min: 0
      max: 100

# This is invalid - no overrides allowed
metrics:
  - value:
      source:
        instance: shared
        max: 200 # ERROR: Cannot override instance
```

### Hierarchical References

Templates can reference other templates:

```yaml
templates:
  clocks:
    base_clock:
      type: periodic
      interval: 1s

  sources:
    base_source:
      type: random_int
      clock:
        template: base_clock # Reference clock template
      min: 0
      max: 100

  values:
    base_value:
      source:
        template: base_source # Reference source template
      transforms: [accumulate]
```

Instances can reference templates or other instances:

```yaml
instances:
  sources:
    shared_source:
      clock:
        template: base_clock # Instance can use template
      # ...

  values:
    shared_value:
      source:
        instance: shared_source # Instance referencing instance
      # ...
```
