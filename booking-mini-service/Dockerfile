FROM rust:latest AS build

WORKDIR /usr/src/pizzeria_carlos

COPY Cargo.toml Cargo.lock ./

RUN mkdir src && echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -f target/release/deps/pizzeria_carlos*

COPY . .

EXPOSE 4567

RUN cargo run
