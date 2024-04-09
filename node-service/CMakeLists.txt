cmake_minimum_required(VERSION 3.27)
project(node_service)

set(CMAKE_CXX_STANDARD 17)

file(GLOB SOURCES "src/*.cpp")

add_executable(node_service src/main.cpp)

