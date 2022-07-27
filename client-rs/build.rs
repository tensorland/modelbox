fn main() {
    tonic_build::configure()
        .format(false)
        .compile(&["service.proto"], &["../proto"])
        .unwrap();
}
