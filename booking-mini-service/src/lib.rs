mod adapter;
mod application;
mod domain;
mod infrastructure;

#[macro_use]
extern crate derive_builder;

#[macro_use]
extern crate shaku;

#[cfg(test)]
#[macro_use]
extern crate lazy_static;