use tauri_plugin_shell::ShellExt;
use tauri_plugin_shell::process::CommandEvent;

#[tauri::command]
async fn run_network_scan(app: tauri::AppHandle, subnet: String) -> Result<String, String> {
    // 1. Call the sidecar binary
    let sidecar_command = app.shell()
        .sidecar("network-engine")
        .map_err(|e| e.to_string())?
        .args(["--subnet", &subnet]);

    // 2. Execute and capture output
    let output = sidecar_command.output().await.map_err(|e| e.to_string())?;

    if output.status.success() {
        Ok(String::from_utf8(output.stdout).unwrap_or_else(|_| "Invalid UTF8".into()))
    } else {
        Err(String::from_utf8(output.stderr).unwrap_or_else(|_| "Unknown Error".into()))
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![run_network_scan])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}