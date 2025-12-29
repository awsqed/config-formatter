# Docker Compose Formatter

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/awsqed/docker-compose-formatter)](https://goreportcard.com/report/github.com/awsqed/docker-compose-formatter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/awsqed/docker-compose-formatter)](https://golang.org/)

A CLI tool for formatting Docker Compose files with consistent indentation and directive ordering.

## Features

- **Consistent Indentation**: Configurable space-based indentation (default: 2 spaces)
- **Consistent Directive Order**: Automatically orders directives according to Docker Compose best practices
  - Top-level: `version`, `name`, `networks`, `volumes`, `configs`, `secrets`, `services` (services last)
  - Service-level: `image`, `build`, `container_name`, `environment`, `ports`, `volumes`, `depends_on`, etc.
- **Service Separation**: Automatically adds empty lines between services for better readability
- **Comment Preservation**: Comments are preserved in their original positions
- **Multiple Output Options**: Print to stdout, write to file, or modify in-place
- **Format Checking**: Verify if files are already formatted

## Installation

```bash
go install github.com/awsqed/docker-compose-formatter@latest
```

Or build from source:

```bash
git clone https://github.com/awsqed/docker-compose-formatter
cd docker-compose-formatter
go build -o docker-compose-formatter
```

## Usage

### Basic Usage

Format a file and print to stdout:
```bash
docker-compose-formatter -input docker-compose.yml
```

### Write to Output File

```bash
docker-compose-formatter -input docker-compose.yml -output formatted-docker-compose.yml
```

### Format In-Place

```bash
docker-compose-formatter -input docker-compose.yml -w
```

### Custom Indentation

```bash
docker-compose-formatter -input docker-compose.yml -indent 4
```

### Check if File is Formatted

```bash
docker-compose-formatter -input docker-compose.yml -check
```

This will exit with code 0 if the file is formatted, or 1 if it needs formatting.

## Command-Line Flags

- `-input` (required): Input docker-compose file path
- `-output`: Output file path (if not specified, prints to stdout)
- `-w`: Write result to source file instead of stdout
- `-indent`: Number of spaces for indentation (default: 2)
- `-check`: Check if file is formatted without making changes

## Directive Ordering

The formatter applies consistent ordering to directives:

### Top-Level Directives
1. `version`
2. `name`
3. `networks`
4. `volumes`
5. `configs`
6. `secrets`
7. `services` (placed last as it's typically the longest section)

### Service-Level Directives
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

Keys not in the predefined order are sorted alphabetically within their group.

## Example

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
  api:
    image: node:18
    ports:
      - "3000:3000"
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

  api:
    image: node:18
    ports:
      - "3000:3000"
```

Note the improvements:
- Top-level sections reordered (networks before services)
- Services have proper directive ordering
- Empty line added between services
- Consistent indentation throughout

## Development

### Running Without Building

Use the Makefile to run without building:
```bash
make run FILE=tailscale.yml
make run FILE=vaultwarden.yml ARGS="-check"
make run FILE=example-docker-compose.yml ARGS="-w"
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

Run with the included example file:
```bash
go run main.go -input example-docker-compose.yml
```

Run tests:
```bash
make test
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

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
