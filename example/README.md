# Example C++ Project

A simple C++ project demonstrating cppenv with CMake and Conan.

## Setup

```bash
# Install cppenv tools
cppenv install

# Install Conan dependencies
cppenv run setup

# Configure with CMake
cppenv run configure

# Build
cppenv run build

# Run
cppenv run run
```

Or do it all at once:

```bash
cppenv install
cppenv run all
cppenv run run
```

## Project Structure

```
example/
├── CMakeLists.txt      # CMake build configuration
├── conanfile.txt       # Conan dependencies (fmt library)
├── cppenv.toml         # cppenv tool versions and scripts
├── README.md
└── src/
    └── main.cpp        # Application source
```
