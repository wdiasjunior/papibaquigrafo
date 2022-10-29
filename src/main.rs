#![allow(non_snake_case)]

extern crate reqwest;
extern crate tokio;

// use std::fs;
// use std::fs::File;
// use std::{io, thread, time};

mod mangadex;

#[tokio::main]
async fn main() {

  let client = reqwest::blocking::Client::new();

  mangadex::getManga(&client).await;
}
