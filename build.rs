use std::env;
use std::fs;
use std::path::Path;

fn collect_yaml(path: &Path, files: &mut Vec<String>) {
    for entry in fs::read_dir(path).expect("read calendar directory") {
        let path = entry.expect("read calendar entry").path();
        if path.is_dir() {
            collect_yaml(&path, files);
        } else if path.extension().is_some_and(|ext| ext == "yaml") {
            files.push(
                path.to_str()
                    .expect("UTF-8 calendar path")
                    .replace('\\', "/"),
            );
        }
    }
}

fn main() {
    println!("cargo:rerun-if-changed=data");
    let mut files = Vec::new();
    collect_yaml(Path::new("data/markets"), &mut files);
    collect_yaml(Path::new("data/holidays"), &mut files);
    files.sort();
    let mut source = String::from("const FILES: &[(&str, &str)] = &[\n");
    for path in files {
        source.push_str(&format!(
            "({path:?}, include_str!(concat!(env!(\"CARGO_MANIFEST_DIR\"), \"/\", {path:?}))),\n"
        ));
    }
    source.push_str("];\n");
    fs::write(
        Path::new(&env::var_os("OUT_DIR").expect("Cargo output directory")).join("calendars.rs"),
        source,
    )
    .expect("write embedded calendar inventory");
}
