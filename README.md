# obsbox

Telemetry signal generator for testing observability components. obsbox generates coherent metric signals derived from the same underlying data sources, enabling verification of monitoring pipelines, collectors, and transformation logic without requiring live systems.

Built on [simv](https://github.com/neox5/simv) for guaranteed mathematical consistency between related metrics.

## Quick Start

```bash
# Run with default configuration
podman run -p 9090:9090 ghcr.io/neox5/obsbox:latest

# Run with custom configuration
podman run -p 9090:9090 \
  -v $(pwd)/custom-config.yaml:/config/config.yaml:ro \
  ghcr.io/neox5/obsbox:latest
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
wget https://github.com/neox5/obsbox/releases/download/v0.2.0/obsbox-linux-amd64
chmod +x obsbox-linux-amd64
./obsbox-linux-amd64 --version
```

Verify checksums:

```bash
wget https://github.com/neox5/obsbox/releases/download/v0.2.0/obsbox-linux-amd64.sha256
sha256sum -c obsbox-linux-amd64.sha256
```

### Build from Source

Requires Go 1.25+:

```bash
go install github.com/neox5/obsbox/cmd/obsbox@latest
```

## Usage

```
obsbox -config <path>    Path to configuration file
obsbox --version         Print version and exit
```

## Configuration Example

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

More advanced example with templates and instances:

```yaml
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

metrics:
  # Using shared instance for coherent counter/gauge pair
  - name: events_total
    type: counter
    description: "Total events"
    value:
      source:
        instance: shared_source
      transforms: [accumulate]

  - name: events_current
    type: gauge
    description: "Recent events"
    value:
      source:
        instance: shared_source
      transforms: [accumulate]
      reset: on_read

  # Using template with override
  - name: requests_total
    type: counter
    description: "Total requests"
    value:
      source:
        template: base_events
        max: 200 # Override template value
      transforms: [accumulate]

export:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
```

See `examples/` directory and documentation for more configuration patterns.

## Documentation

- [Configuration Guide](doc/configuration.md) - Complete configuration guide with examples and patterns
- [Configuration Reference](doc/reference.md) - Detailed parameter reference and specifications

## License

MIT License - see [LICENSE](LICENSE) file for details.
