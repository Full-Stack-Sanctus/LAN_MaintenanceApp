fn main() {
    // Set up custom windows attributes
    let mut windows = tauri_build::WindowsAttributes::new();
    
    // Inject your custom app.manifest file natively
    windows = windows.app_manifest(include_str!("app.manifest"));

    // Apply the attributes to Tauri's build system
    let attrs = tauri_build::Attributes::new().windows_attributes(windows);
    
    tauri_build::try_build(attrs).expect("failed to run tauri build system");
}