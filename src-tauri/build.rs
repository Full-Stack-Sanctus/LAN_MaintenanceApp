fn main() {
    // Only run this block if we are compiling on/for Windows
    #[cfg(target_os = "windows")]
    {
        let mut res = winres::WindowsResource::new();
        res.set_manifest_file("app.manifest");
        res.compile().unwrap();
    }

    tauri_build::build();
}