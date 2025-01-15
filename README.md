# Port Watch

A desktop application for monitoring and managing active ports and processes on your system.

## Key Features

- Real-time port monitoring
- Process search by name
- Process termination capability (with system-critical process protection)
- Automatic refresh every 5 seconds
- Intuitive user interface

## Tech Stack

- Go language
- Fyne GUI framework (v2)

## Installation

1. Install Go 1.16 or higher
2. Clone the project

```bash
git clone https://github.com/Kim-DaeHan/port-watch.git
cd port-watch
```

3. Install dependencies

```bash
go mod download
```

4. Run the application

```bash
go run main.go
```

## Build

Build for macOS:

```bash
fyne package -os darwin -icon icon.icns
```

## License

MIT License

## Developer

- Kim-DaeHan
