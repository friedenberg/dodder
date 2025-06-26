use super::digest::Digest;
use crypto::digest::Digest as OOPS_THEIR_DGIEST;
use crypto::sha2::Sha256;
use std::io::{Error, Write};

pub struct Writer(pub Sha256);

impl Writer {
    pub fn new() -> Self {
        return Self(Sha256::new());
    }

    pub fn digest(&mut self) -> Digest {
        let mut b = [0 as u8; 32];
        self.0.result(&mut b[..]);
        return Digest::new(b);
    }
}

impl Write for Writer {
    fn write(&mut self, buf: &[u8]) -> Result<usize, Error> {
        self.0.input(buf);
        Ok(buf.len())
    }

    fn flush(&mut self) -> Result<(), Error> {
        Ok(())
    }
}
