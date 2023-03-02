use std::path::PathBuf;

use clap::{Args, Parser, Subcommand};
use server_config::ServerConfig;
mod agent;
mod grpc_server;
mod model_helper;
mod modelbox;
mod server_config;
mod repository;

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let subscriber = tracing_subscriber::FmtSubscriber::new();
    tracing::subscriber::set_global_default(subscriber).unwrap();

    if cli.config_path.is_none() {
        panic!("config path is required")
    }

    match cli.command {
        Commands::Server(server) => match server.commands {
            StartCommands::Start => {
                start_agent(cli.config_path.unwrap()).await;
            }
            StartCommands::InitConfig => {
                ServerConfig::generate_config(cli.config_path.unwrap())
                    .unwrap_or_else(|e| panic!("unable to write config {}", e));
            }
        },
    }
}

async fn start_agent(config_path: PathBuf) {
    let config = ServerConfig::from_path(config_path)
        .unwrap_or_else(|e| panic!("unable to read config {}", e));
    let agent = agent::Agent::new(config).await;
    tokio::select! {
        _ = agent.start() => {
            println!("agent has stopped runnning")
        }
        _ = agent.wait_for_signal() => {}
    }
}

#[derive(Debug, Parser)]
#[command(about = "tensorland cli", long_about = None)]
struct Cli {
    #[arg(global = true)]
    config_path: Option<PathBuf>,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Debug, Subcommand)]
enum Commands {
    Server(StartArgs),
}

#[derive(Debug, Args)]
struct StartArgs {
    #[command(subcommand)]
    commands: StartCommands,
}

#[derive(Debug, Subcommand)]
enum StartCommands {
    Start,
    InitConfig,
}
