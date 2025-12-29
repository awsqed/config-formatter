# Config Formatter

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A modular CLI tool for formatting YAML configuration files with consistent indentation and directive ordering. Currently supports Docker Compose and Traefik configurations.

## Features

- **Multi-Format Support**: Automatically detects and formats different config types
  - Docker Compose files
  - Traefik configuration files
  - Extensible architecture for adding more formats
- **Auto-Detection**: Automatically identifies config type based on filename and content
- **Consistent Indentation**: Configurable space-based indentation (default: 2 spaces)
- **Smart Directive Ordering**: Format-specific ordering rules for better readability
- **Comment Preservation**: Comments are preserved in their original positions
- **Multiple Output Options**: Print to stdout, write to file, or modify in-place
- **Format Checking**: Verify if files are already formatted

## Installation

```bash
go install github.com/awsqed/config-formatter@latest
```

Or build from source:

```bash
git clone https://github.com/awsqed/config-formatter
cd config-formatter
go build -o config-formatter
```

## Usage

### Basic Usage

Format a file and print to stdout (auto-detects format):
```bash
config-formatter -input docker-compose.yml
config-formatter -input traefik.yml
```

### Specify Format Type

```bash
config-formatter -input myfile.yml -type docker-compose
config-formatter -input myfile.yml -type traefik
```

### Write to Output File

```bash
config-formatter -input docker-compose.yml -output formatted.yml
```

### Format In-Place

```bash
config-formatter -input docker-compose.yml -w
```

### Custom Indentation

```bash
config-formatter -input traefik.yml -indent 4
```

### Check if File is Formatted

```bash
config-formatter -input docker-compose.yml -check
```

This will exit with code 0 if the file is formatted, or 1 if it needs formatting.

## Command-Line Flags

- `-input` (required): Input config file path
- `-output`: Output file path (if not specified, prints to stdout)
- `-w`: Write result to source file instead of stdout
- `-indent`: Number of spaces for indentation (default: 2)
- `-check`: Check if file is formatted without making changes
- `-type`: Formatter type to use (`docker-compose`, `traefik`). Auto-detected if not specified

## Supported Formats

### Docker Compose

Formats Docker Compose files with best-practice directive ordering.

**Top-Level Directives:**
1. `version`
2. `name`
3. `networks`
4. `volumes`
5. `configs`
6. `secrets`
7. `services` (placed last as it's typically the longest section)

**Service-Level Directives:**
1. `image`
2. `build`
3. `container_name`
4. `hostname`
5. `command`
6. `entrypoint`
7. `environment`
8. `env_file`
9. `ports`
10. `expose`
11. `volumes`
12. `networks`
13. `depends_on`
14. And many more...

**Example:**

Before formatting:
```yaml
services:
  web:
    ports:
      - "8080:80"
    image: nginx:latest
    volumes:
      - ./html:/usr/share/nginx/html
    restart: always
networks:
  default:
```

After formatting:
```yaml
networks:
  default:

services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    volumes:
      - ./html:/usr/share/nginx/html
    restart: always
```

### Traefik

Formats Traefik configuration files with logical grouping and ordering.

**Top-Level Directives:**
1. Global configuration (`log`, `api`, `metrics`, etc.)
2. `entryPoints`
3. `providers`
4. `certificatesResolvers`
5. Protocol sections (`http`, `tcp`, `udp`, `tls`)

**HTTP Section:**
1. `routers`
2. `services`
3. `middlewares`
4. `serversTransports`

Keys not in the predefined order are sorted alphabetically within their group.

## Architecture

The formatter uses a modular plugin architecture:

- `formatter/formatter.go`: Core interface and base functionality
- `modules/dockercompose/`: Docker Compose formatter implementation
- `modules/traefik/`: Traefik formatter implementation

### Adding New Formatters

To add support for a new config format:

1. Create a new module directory under `modules/`
2. Implement the `Formatter` interface:
   - `Format(data []byte, indent int) ([]byte, error)` - Format the config
   - `Name() string` - Return formatter name
   - `CanHandle(filename string, data []byte) bool` - Detect if file matches this format
3. Register the formatter in `main.go`

## Development

### Running Without Building

```bash
go run main.go -input compose.yml
go run main.go -input traefik.yml -type traefik
```

### Building for Multiple Platforms

Build for all platforms:
```bash
make build-all
```

Build for specific platforms:
```bash
make build-linux-amd64
make build-darwin-arm64
make build-windows-amd64
```

See all available commands:
```bash
make help
```

### Testing

Run tests:
```bash
make test
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

To contribute a new formatter module:
1. Implement the `Formatter` interface
2. Add detection logic in `CanHandle()`
3. Define format-specific ordering rules
4. Update this README with format documentation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 awsqed

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
