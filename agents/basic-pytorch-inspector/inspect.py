from abc import ABC, abstractmethod

from modelbox.modelbox import Event, Model, Experiment

class ModelBoxAgent(ABC):

    @abstractmethod
    def handle_event(event: Event):
        pass

    def model(self) -> Model:
        pass

    def experiment(self) -> Experiment:
        pass


class PytorchModelInspectorAgent(ModelBoxAgent):

    def __init__(self) -> None:
        super().__init__()

    def handle_event(event: Event):
        pass

    def inspect_model(model):
        model.to_onnx()
        model.get_num_params()
        pass

if __name__ == "__main__":
    print("hello world")