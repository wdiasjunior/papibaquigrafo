use std::io;

pub fn funcHandleInput() {
  let mut tasks: Vec<String> = Vec::new();
  let mut i: usize = 0;
  loop {
    println!("add a task: ");

    let mut input: String = String::new();
    io::stdin().read_line(&mut input).expect("Couldn’t read from stdin");

    if input.trim() == "exit" {
      break;
    }

    tasks.push(input);
    println!("{}", tasks[i]);
    
    i += 1;
  }
}
