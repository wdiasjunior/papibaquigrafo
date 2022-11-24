// some tui tests. this lib's docs sucks. how the fuck are people use this shit?
// disables warnings while figuring out the TUI stuff
#![allow(unused_variables, unused_mut, dead_code)]
#![cfg_attr(debug_assertions, allow(dead_code, unused_imports))]

use crossterm::{
  event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode},
  execute,
  terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use std::{
  error::Error,
  io,
  time::{Duration, Instant},
};
use tui::{
  backend::{Backend, CrosstermBackend},
  layout::{Constraint, Corner, Direction, Layout},
  style::{Color, Modifier, Style},
  text::{Span, Spans},
  widgets::{Block, Borders, List, ListItem, ListState},
  Frame, Terminal,
};

struct StatefulList<T> {
  state: ListState,
  items: Vec<T>,
}

impl<T> StatefulList<T> {
  fn with_items(items: Vec<T>) -> StatefulList<T> {
    StatefulList {
      state: ListState::default(),
      items,
    }
  }

  fn next(&mut self) {
    let i = match self.state.selected() {
      Some(i) => {
        if i >= self.items.len() - 1 {
          0
        } else {
          i + 1
        }
      }
      None => 0,
    };
    self.state.select(Some(i));
  }

  fn previous(&mut self) {
    let i = match self.state.selected() {
      Some(i) => {
        if i == 0 {
          self.items.len() - 1
        } else {
          i - 1
        }
      }
      None => 0,
    };
    self.state.select(Some(i));
  }

  fn unselect(&mut self) {
    self.state.select(None);
  }

  fn select(&mut self) {
    self.state.select(Some(0));
  }

  fn selectSource(&mut self) {
    println!("selectSource");
  }
}

struct App<'a> {
  mangaSourcesItems: StatefulList<(&'a str, usize)>,
}

impl<'a> App<'a> {
  fn new() -> App<'a> {
    App {
      mangaSourcesItems: StatefulList::with_items(vec![
        ("Mangadex", 1),
        ("TCB Scans", 1)
      ]),
    }
  }
}

fn main() -> Result<(), Box<dyn Error>> {
  // setup terminal
  enable_raw_mode()?;
  let mut stdout = io::stdout();
  execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
  let backend = CrosstermBackend::new(stdout);
  let mut terminal = Terminal::new(backend)?;

  // create app and run it
  let tick_rate = Duration::from_millis(250);
  let app = App::new();
  let res = runApp(&mut terminal, app, tick_rate);

  // restore terminal
  disable_raw_mode()?;
  execute!(
      terminal.backend_mut(),
      LeaveAlternateScreen,
      DisableMouseCapture
  )?;
  terminal.show_cursor()?;

  if let Err(err) = res {
    println!("{:?}", err)
  }

  return Ok(());
}

fn runApp<B: Backend>(terminal: &mut Terminal<B>, mut app: App, tick_rate: Duration) -> io::Result<()> {
  let mut last_tick = Instant::now();
  app.mangaSourcesItems.select();
  loop {
    terminal.draw(|f| ui(f, &mut app))?;

    let timeout = tick_rate.checked_sub(last_tick.elapsed()).unwrap_or_else(|| Duration::from_secs(0));
    if crossterm::event::poll(timeout)? {
      if let Event::Key(key) = event::read()? {
        match key.code {
          KeyCode::Char('q') => return Ok(()),
          KeyCode::Left => app.mangaSourcesItems.unselect(),
          KeyCode::Down => app.mangaSourcesItems.next(),
          KeyCode::Up => app.mangaSourcesItems.previous(),
          KeyCode::Enter => app.mangaSourcesItems.selectSource(),
          _ => {}
        }
      }
    }
  }
}

fn ui<B: Backend>(f: &mut Frame<B>, app: &mut App) {
  // Creates the list view
  let chunks = Layout::default()
    .direction(Direction::Horizontal)
    .constraints([Constraint::Percentage(100)])
    .split(f.size());

  // Iterate through all elements in the `items` app
  let mangaSourcesItems: Vec<ListItem> = app.mangaSourcesItems.items.iter()
    .map(|i| {
      let mut lines = vec![Spans::from(i.0)];
      ListItem::new(lines).style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
    })
    .collect();

  // Create a List from all list items and highlight the currently selected one
  let mangaSourcesList = List::new(mangaSourcesItems)
    .block(Block::default().borders(Borders::ALL).title(" papibaquigrafo "))
    .highlight_style(
      Style::default()
      .fg(Color::LightBlue)
      .add_modifier(Modifier::BOLD),
    )
    .highlight_symbol("> ");

  // renders the item list
  f.render_stateful_widget(mangaSourcesList, chunks[0], &mut app.mangaSourcesItems.state);
}
