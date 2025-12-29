# cppenv

Reproducible C++ build environments using pip-installable tools. Think npm/cargo for C++ toolchains.

## Problem

C++ developers face constant friction with toolchain inconsistencies:
- "Works on my machine" due to different CMake/compiler versions
- CI environments differ from local dev machines
- Onboarding new developers requires manual tool installation
- No standard way to pin toolchain versions per-project

## Solution

A single Go binary that:
1. Manages its own Python interpreter (downloaded on first run)
2. Reads a `cppenv.toml` config file declaring tool versions
3. Creates an isolated virtualenv per project
4. Pip installs all tools into that venv
5. Provides a `run` command to execute tools or user-defined scripts

## Installation

Download the latest release for your platform from [Releases](https://github.com/michxymi/cppenv/releases).

Or build from source:

```bash
go build -o cppenv .
```

## Quick Start

```bash
# Initialize a new project (fetches latest tool versions from PyPI)
cd my-cpp-project
cppenv init

# Install all tools (downloads Python on first run)
cppenv install

# Use tools directly
cppenv run cmake --version
cppenv run conan install . --build=missing
cppenv run cmake -B build -G Ninja

# Or define scripts in cppenv.toml and run them
cppenv run build
```

## Configuration

`cppenv.toml`:

```toml
[project]
name = "my-project"

[tools]
ziglang = "0.13.0"
cmake = "3.29.2"
ninja = "1.11.1.1"
conan = "2.3.0"
clang-tools = "18.1.3"

[scripts]
build = "cmake --build build"
test = "ctest --test-dir build"
fmt = "clang-format -i src/*.cpp"
```

## Commands

| Command | Description |
|---------|-------------|
| `cppenv init` | Create cppenv.toml with latest tool versions |
| `cppenv install` | Download Python (if needed) and install tools |
| `cppenv run <cmd>` | Run a command or script with tools in PATH |
| `cppenv status` | Show project info and installed tools |
| `cppenv toolchain` | Regenerate CMake toolchain file for Zig |

## Default Tools

| Package | Purpose |
|---------|---------|
| ziglang | C/C++ compiler (Clang/LLVM via Zig) |
| cmake | Build generator |
| ninja | Build tool |
| conan | Package manager |
| clang-tools | clang-format, clang-tidy |

## CMake Integration

After `cppenv install`, use the generated toolchain file:

```bash
cppenv run cmake -B build -G Ninja \
  -DCMAKE_TOOLCHAIN_FILE=.cppenv/zig-toolchain.cmake
```

## License

MIT
