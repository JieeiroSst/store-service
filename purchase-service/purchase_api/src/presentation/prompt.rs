use std::io::{BufRead, Write};

pub fn read_input<R>(mut input: R) -> String
where
    R: BufRead,
{
    let mut input_text: String = String::new();

    input
        .read_line(&mut input_text)
        .expect("Failed to read input");

    input_text.trim().to_string()
}

pub fn ask_question<R, W>(input: R, mut output: W, question: &str) -> String
where
    R: BufRead,
    W: Write,
{
    writeln!(output, "{}", question).unwrap();

    read_input(input)
}

pub fn menu<R, W, E>(input: R, mut output: W, mut error_output: E, options: Vec<&str>) -> u8
where
    R: BufRead,
    W: Write,
    E: Write,
{
    writeln!(output, "\nMENU").unwrap();
    writeln!(output, "----------------------------------------\n").unwrap();
    writeln!(output, "Please, select an option:").unwrap();

    let mut i = 1;

    for option in options {
        writeln!(output, "\t {}. {}", i, option).unwrap();
        i += 1;
    }

    writeln!(output, "\t 0. Exit").unwrap();
    write!(output, "\nOption: ").unwrap();

    output.flush().expect("Error flushing");

    let option = read_input(input).trim().parse();

    writeln!(output, "\n").unwrap();

    match option {
        Ok(o) => o,
        Err(_) => {
            writeln!(error_output, "ERROR: Please, type a number").unwrap();
            u8::MAX
        }
    }
}