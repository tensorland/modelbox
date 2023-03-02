fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = "./src";

    tonic_build::configure()
        .out_dir(out_dir)
        .compile(&["../proto/service.proto"], &["../proto"])
        .unwrap_or_else(|e| panic!("Failed to compile protos {:?}", e));

    Ok(())
}
