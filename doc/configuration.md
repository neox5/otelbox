# Configuration Guide

Complete guide to configuring obsbox for telemetry signal generation.

## Configuration File Structure

obsbox uses YAML configuration files with five main sections:

- **templates** - Reusable definitions with override support
- **instances** - Named, shared objects
- **metrics** - Maps values to exposed metrics
- **export** - Configures metric exposition (Prometheus, OTEL)
- **settings** - Application settings

## Simple Example

Minimal configuration generating a single counter metric:

```yaml
instances:
  clocks:
    tick:
      type: periodic
      interval: 1s

  sources:
    events:
      type: random_int
      clock:
        instance: tick
      min: 0
      max: 10

  values:
    total_events:
      source:
        instance: events
      transforms: [accumulate]

metrics:
  - name: app_events_total
    type: counter
    description: "Total events processed"
    value:
      instance: total_events
    attributes:
      service: myapp

export:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
```

## Configuration Sections

### Templates & Instances

Templates and instances provide reusable configuration definitions with different characteristics:

**Templates** are reusable definitions that can be referenced with overrides:

- Define once, reference many times with variations
- Override specific fields when referenced
- Useful for creating variations on common patterns
- Example: Base source template with different min/max values per metric

**Instances** are concrete, named objects that can be shared across references:

- Reference by name without modification
- Guarantee identical behavior across all references
- Useful for shared clocks or sources
- Example: Single clock instance driving multiple sources

#### Templates Section

Templates support hierarchical definitions - templates can reference other templates:

```yaml
templates:
  clocks:
    tick_1s:
      type: periodic
      interval: 1s

    tick_fast:
      type: periodic
      interval: 100ms

  sources:
    base_events:
      type: random_int
      clock:
        template: tick_1s
      min: 0
      max: 100

  values:
    counter_value:
      source:
        template: base_events
      transforms: [accumulate]

  metrics:
    counter_metric:
      type: counter
      value:
        template: counter_value
```

**Template References with Overrides:**
Templates can be referenced and specific fields overridden:

```yaml
metrics:
  - name: events_total
    type: counter
    description: "Total events"
    value:
      template: counter_value
      source:
        template: base_events
        max: 50 # Override max from template
    attributes:
      service: myapp
```

#### Instances Section

Instances define concrete, shareable objects:

```yaml
instances:
  clocks:
    main_tick:
      type: periodic
      interval: 1s

    fast_tick:
      type: periodic
      interval: 100ms

  sources:
    event_source:
      type: random_int
      clock:
        instance: main_tick # Reference clock instance
      min: 0
      max: 100

  values:
    total_events:
      source:
        instance: event_source # Reference source instance
      transforms: [accumulate]
```

**Instance References (No Overrides):**
Instances are referenced by name and cannot be modified:

```yaml
metrics:
  - name: events_total
    type: counter
    description: "Total events"
    value:
      instance: total_events # Reference value instance
    attributes:
      service: myapp
```

#### Reference Types

Configuration fields that reference other objects support three forms:

**1. Instance Reference:**

```yaml
value:
  instance: total_events # Reference named instance
```

**2. Template Reference (with optional overrides):**

```yaml
value:
  template: counter_value # Reference template
  reset: on_read # Override reset behavior
```

**3. Inline Definition:**

```yaml
value:
  source:
    type: random_int
    clock:
      type: periodic
      interval: 1s
    min: 0
    max: 100
  transforms: [accumulate]
```

#### Seed

Optional seed configuration for reproducible simulations:

```yaml
settings:
  seed: 12345
```

When omitted, a time-based seed is used (logged at startup for reproduction).

### Metrics Definition

Maps values to exposed metrics:

```yaml
metrics:
  - name: <metric_name>              # Simple form (same name for both protocols)
    name:                             # Full form (protocol-specific names)
      prometheus: <prom_name>
      otel: <otel_name>
    type: counter | gauge
    description: <help_text>
    value: <value_reference>          # Instance/template reference or inline
    attributes:                       # Labels (Prometheus) / Attributes (OTEL)
      <key>: <value>
```

**Metric Types:**

- `counter` - Monotonically increasing value
- `gauge` - Value that can increase or decrease

**Naming:**

- Simple form: Same name for Prometheus and OTEL
- Full form: Protocol-specific names (e.g., `app.events.total` vs `app_events_total`)

**Value References:**

Metrics reference values in three ways:

1. **Instance reference** - Points to named instance
2. **Template reference** - Uses template with optional overrides
3. **Inline definition** - Complete value definition in metric

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

#### Seed

Optional seed for reproducible simulations:

```yaml
settings:
  seed: <uint64>
```

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

## Complete Example

Full-featured configuration showing templates, instances, and inline definitions:

```yaml
settings:
  seed: 12345

templates:
  clocks:
    tick_1s:
      type: periodic
      interval: 1s

  sources:
    base_events:
      type: random_int
      clock:
        template: tick_1s
      min: 0
      max: 100

  values:
    counter_value:
      source:
        template: base_events
      transforms: [accumulate]

instances:
  clocks:
    main_tick:
      type: periodic
      interval: 1s

  sources:
    shared_source:
      type: random_int
      clock:
        instance: main_tick
      min: 0
      max: 50

  values:
    shared_counter:
      source:
        instance: shared_source
      transforms: [accumulate]

metrics:
  # Metric using instance
  - name:
      prometheus: app_events_total
      otel: app.events.total
    type: counter
    description: "Total events from shared instance"
    value:
      instance: shared_counter
    attributes:
      source: instance

  # Metric using template with overrides
  - name:
      prometheus: app_requests_total
      otel: app.requests.total
    type: counter
    description: "Total requests from template"
    value:
      template: counter_value
      source:
        template: base_events
        max: 200  # Override max
    attributes:
      source: template

  # Metric with inline definition
  - name:
      prometheus: app_errors_total
      otel: app.errors.total
    type: counter
    description: "Total errors from inline definition"
    value:
      source:
        type: random_int
        clock:
          type: periodic
          interval: 500ms
        min: 0
        max: 10
      transforms: [accumulate]
    attributes:
      source: inline

export:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics

  otel:
    enabled: false
    transport: grpc
    host: localhost
    port: 4317
    interval: 10s
    resource:
      service.name: obsbox
      service.version: 0.3.0
      deployment.environment: development

settings:
  internal_metrics:
    enabled: true
    format: native
```

## Common Patterns

### Template with Variations

Define a base template and create variations with overrides:

```yaml
templates:
  sources:
    base_load:
      type: random_int
      clock:
        type: periodic
        interval: 1s
      min: 0
      max: 100

metrics:
  - name: low_load_total
    type: counter
    value:
      source:
        template: base_load
        max: 50 # Low load variant
      transforms: [accumulate]

  - name: high_load_total
    type: counter
    value:
      source:
        template: base_load
        max: 200 # High load variant
      transforms: [accumulate]
```

### Shared Clock Pattern

Multiple sources share a single clock instance:

```yaml
instances:
  clocks:
    main_tick:
      type: periodic
      interval: 1s

  sources:
    source_a:
      type: random_int
      clock:
        instance: main_tick # Shared clock
      min: 0
      max: 100

    source_b:
      type: random_int
      clock:
        instance: main_tick # Same clock instance
      min: 0
      max: 50
```

### Coherent Counter/Gauge Pairs

Create counter and gauge from same source for mathematical coherence:

```yaml
instances:
  sources:
    events:
      type: random_int
      clock:
        type: periodic
        interval: 1s
      min: 0
      max: 100

  values:
    total_events:
      source:
        instance: events
      transforms: [accumulate]

    recent_events:
      source:
        instance: events # Same source guarantees coherence
      transforms: [accumulate]
      reset: on_read

metrics:
  - name: events_total
    type: counter
    value:
      instance: total_events

  - name: events_current
    type: gauge
    value:
      instance: recent_events
```

### Multiple Independent Update Frequencies

Different metric groups with independent clocks:

```yaml
instances:
  clocks:
    fast_tick:
      type: periodic
      interval: 100ms

    slow_tick:
      type: periodic
      interval: 5s

  sources:
    api_requests:
      type: random_int
      clock:
        instance: fast_tick
      min: 0
      max: 100

    batch_jobs:
      type: random_int
      clock:
        instance: slow_tick
      min: 0
      max: 10
```

### Inline Definitions for One-Off Metrics

Use inline definitions when metric is unique and won't be reused:

```yaml
metrics:
  - name: unique_metric_total
    type: counter
    description: "One-off metric"
    value:
      source:
        type: random_int
        clock:
          type: periodic
          interval: 2s
        min: 0
        max: 25
      transforms: [accumulate]
```
