FROM rust:1.61.0
 
WORKDIR /usr/src/shop_service
COPY . .

EXPOSE 6767
 
RUN cargo install --path .
 
CMD ["shop_service"]