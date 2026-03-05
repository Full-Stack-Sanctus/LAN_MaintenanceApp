use tauri_plugin_shell::ShellExt;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct ScanArgs {
    pub target: String,
    pub community: Option<String>,
}

#[tauri::command]
pub async fn execute_enterprise_audit(
    app: tauri::AppHandle,
    args: ScanArgs,
) -> Result<String, String> {
    let mut command_args = vec!["--target", &args.target];

    if let Some(comm) = args.community {
        if !comm.is_empty() {
            command_args.push("--community");
            command_args.push(&comm);
        }
    }

    let output = app
        .shell()
        .sidecar("network-engine")
        .map_err(|e| format!("Sidecar error: {}", e))?
        .args(command_args)
        .output()
        .await
        .map_err(|e| format!("Execution error: {}", e))?;

    if output.status.success() {
        Ok(String::from_utf8_lossy(&output.stdout).to_string())
    } else {
        Err(String::from_utf8_lossy(&output.stderr).to_string())
    }
}