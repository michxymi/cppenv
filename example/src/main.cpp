#include <fmt/core.h>

int main() {
  fmt::print("Hello from cppenv!\n");
  fmt::print("Build tools managed by cppenv:\n");
  fmt::print("  - CMake\n");
  fmt::print("  - Ninja\n");
  fmt::print("  - Conan\n");
  fmt::print("  - Zig (C/C++ compiler)\n");
  return 0;
}
