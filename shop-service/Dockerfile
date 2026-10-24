FROM rust:1.97

RUN cargo install cargo-watch
RUN cargo install diesel_cli --no-default-features --features postgres

WORKDIR /usr/src/app

CMD ["cargo", "watch", "-x", "run"]
