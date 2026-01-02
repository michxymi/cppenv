cmake_minimum_required(VERSION 3.15)
project(example LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(fmt REQUIRED)

add_executable(example src/main.cpp)
target_link_libraries(example PRIVATE fmt::fmt)
