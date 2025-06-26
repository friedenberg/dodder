use anyhow::Error;
use std::io::{BufWriter, Read, Write};

use flate2::write::DeflateEncoder;

type Result<T> = std::result::Result<T, Error>;

pub trait WriteFinish: Write {
    fn finish(self: Box<Self>) -> Result<()> {
        Ok(())
    }
}

impl<T: Write> WriteFinish for BufWriter<T> {}
impl<T: Write> WriteFinish for DeflateEncoder<T> {}

pub trait WrapIO {
    fn wrap_input(&self, reader: Box<dyn Read>) -> Result<Box<dyn Read>>;
    fn wrap_output(&self, writer: Box<dyn WriteFinish>) -> Result<Box<dyn WriteFinish>>;
}

pub struct NopWrapIO;

impl WrapIO for NopWrapIO {
    fn wrap_input(&self, reader: Box<dyn Read>) -> Result<Box<dyn Read>> {
        Ok(reader)
    }

    fn wrap_output(&self, writer: Box<dyn WriteFinish>) -> Result<Box<dyn WriteFinish>> {
        Ok(writer)
    }
}
