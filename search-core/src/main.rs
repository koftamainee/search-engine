use tracing::info;

fn main() {
    tracing_subscriber::fmt::init();
    info!("search-core is running");
}
