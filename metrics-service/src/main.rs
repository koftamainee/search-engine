use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpListener;
use tracing::{info, warn};
use std::time::Instant;

const HEALTH_RESPONSE: &str = "\
HTTP/1.1 200 OK\r\n\
Content-Type: application/json\r\n\
Content-Length: 15\r\n\
\r\n\
{\"status\":\"ok\"}";

const NOT_FOUND: &str = "\
HTTP/1.1 404 Not Found\r\n\
Content-Length: 0\r\n\
\r\n";

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| {
                eprintln!("RUST_LOG is not defined, using info level");
                tracing_subscriber::EnvFilter::new("info")
            }),
        )
        .init();

    let addr = match std::env::var("METRICS_SERVICE_PORT") {
        Ok(port) => format!("0.0.0.0:{port}"),
        Err(_) => {
            warn!("METRICS_SERVICE_PORT is not defined, using 8080 port");
            "0.0.0.0:8080".into()
        }
    };

    info!("metrics service listening on {addr}");

    let listener = TcpListener::bind(addr).await?;

    loop {
        let (mut stream, peer) = listener.accept().await?;

        tokio::spawn(async move {
            let start = Instant::now();
            let mut buf = [0u8; 1024];
            let n = match stream.read(&mut buf).await {
                Ok(n) if n > 0 => n,
                _ => return,
            };

            let request = &buf[..n];
            let request_str = String::from_utf8_lossy(request);
            let (method, route) = parse_request_line(&request_str);
            let user_agent = extract_user_agent(request);

            let (status_code, response) = if request.starts_with(b"GET /health") {
                (200, HEALTH_RESPONSE)
            } else {
                (404, NOT_FOUND)
            };

            let _ = stream.write_all(response.as_ref()).await;
            let _ = stream.shutdown().await;

            let elapsed = start.elapsed();

            info!(
                method = method,
                host = "localhost",
                route = route,
                user_agent = user_agent,
                status_code = status_code,
                peer = %peer,
                busy = format!("{:.0}µs", elapsed.as_micros()),
                "HTTP request"
            );
        });
    }
}

fn parse_request_line(request: &str) -> (&str, &str) {
    let parts: Vec<&str> = request.split_whitespace().collect();
    if parts.len() >= 2 {
        (parts[0], parts[1])
    } else {
        ("UNKNOWN", "/")
    }
}

fn extract_user_agent(request: &[u8]) -> String {
    let request_str = String::from_utf8_lossy(request);
    for line in request_str.lines() {
        if line.starts_with("User-Agent:") {
            return line.trim_start_matches("User-Agent:").trim().to_string();
        }
    }
    "-".to_string()
}