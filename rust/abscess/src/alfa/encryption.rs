use crate::prelude::*;
use serde::{Deserialize, Serialize};
use std::fmt::{self, Display, Formatter};
use std::path::PathBuf;

use super::age::Age;
use super::wrap_io::{NopWrapIO, WrapIO};

#[derive(Serialize, Deserialize, Debug, Clone, clap::ValueEnum)]
pub enum Type {
    None,
    Age,
}

impl Display for Type {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        match self {
            Self::None => write!(f, "none"),
            Self::Age => write!(f, "age"),
        }
    }
}

impl Default for Type {
    fn default() -> Self {
        Type::None
    }
}
