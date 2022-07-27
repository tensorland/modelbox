mod proto {
    tonic::include_proto!("modelbox");
}

pub struct ClientConfig {
    pub server_addr: String
}

impl ClientConfig {
    pub fn new(_path: String) -> Box<ClientConfig> {
        Box::new(ClientConfig{server_addr: ":8085".to_owned()})
    }
}

pub struct ModelBoxClient {

}

impl ModelBoxClient {
    pub fn new() -> Self{
        ModelBoxClient{}
    }
}

#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        let result = 2 + 2;
        assert_eq!(result, 4);
    }
}
