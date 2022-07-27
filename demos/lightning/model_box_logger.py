import os
import logging
from re import A
from typing import Optional, Mapping, Callable, Sequence
from unicodedata import name

from pytorch_lightning.loggers.base import LightningLoggerBase, rank_zero_experiment
from pytorch_lightning.utilities.distributed import rank_zero_only

from weakref import ReferenceType
from pytorch_lightning.callbacks.model_checkpoint import ModelCheckpoint

import sys

sys.path.append("../../client-py/modelbox")
from modelbox import ModelBoxClient, Experiment, Project, MLFramework


logger = logging.getLogger("pytorch_lightning")

SERVER_ADDR = "localhost:8085"


class ModelBoxLogger(LightningLoggerBase):
    def __init__(
        self,
        project_name: str,
        experiment_name: str,
        owner: str,
        external_id: str = "",
        upload_checkpoints: bool = False,
        agg_key_funcs: Optional[
            Mapping[str, Callable[[Sequence[float]], float]]
        ] = None,
        agg_default_func: Optional[Callable[[Sequence[float]], float]] = None,
    ):
        self._project_name = project_name
        self._experiment_name = experiment_name
        self._owner = owner
        self._external_id = external_id
        super().__init__(agg_key_funcs=agg_key_funcs, agg_default_func=agg_default_func)

        self._experiment = None

        # Create the MBox client
        self._mbox = ModelBoxClient(SERVER_ADDR)
        self._project = None
        self._experiment = None
        self._upload_checkpoints = upload_checkpoints
        self._checkpoint_paths = set()

        self._current_step = 0
        self._epoch = 0

    @property
    def name(self):
        if self._experiment is None:
            self._experiment = self.experiment
        return self._experiment.name

    @property
    def version(self):
        # Return the experiment version, int or str.
        return "0.1"

    @rank_zero_only
    def log_hyperparams(self, params):
        # params is an argparse.Namespace
        # your code to record hyperparameters goes here
        pass

    @property
    @rank_zero_experiment
    def experiment(self) -> Experiment:
        logger.info("modelbox - attempting to create a project")
        if self._project is None:
            self._project = self._mbox.create_project(self._project_name, self._owner)
        if self._experiment is None:
            self._experiment = self._mbox.create_experiment(
                self._experiment_name,
                self._project.name,
                self._project.owner,
                self._external_id,
                MLFramework.PYTORCH,
            )
        logger.info(
            "modelbox - created experimetn with id: {}".format(self._experiment.id)
        )
        return self._experiment

    @rank_zero_only
    def log_metrics(self, metrics, step):
        self._current_step = step
        self._epoch = metrics["epoch"]
        logger.info(
            "modelbox - log metrics, step: {} metrics: {}".format(step, metrics)
        )

    @rank_zero_only
    def after_save_checkpoint(
        self, checkpoint_callback: "ReferenceType[ModelCheckpoint]"
    ) -> None:
        # Finding out paths of new checkpoints and recording them
        file_names = set()
        chk_state = checkpoint_callback.state_dict()["best_k_models"]
        for best_k_path in chk_state.keys():
            file_names.add(best_k_path)
        new_chk_paths = file_names - self._checkpoint_paths
        for chk_path in new_chk_paths:
            logger.info("modelbox - recording checkpoint {}".format(chk_path))
            metrics = {"val_accu": chk_state[chk_path]}
            chk_id = self._mbox.create_checkpoint_meta(
                self.experiment.id, self._epoch, chk_path, metrics
            )
            logger.info("modelbox - recorded checkpoint {}".format(chk_id))

        # Updating the state with all the checkpoints we have just discovered
        self._checkpoint_paths = file_names

    @rank_zero_only
    def save(self):
        # Optional. Any code necessary to save logger data goes here
        pass

    @rank_zero_only
    def finalize(self, status):
        # Optional. Any code that needs to be run after training
        # finishes goes here
        pass

    @staticmethod
    def _get_full_model_name(
        model_path: str, checkpoint_callback: "ReferenceType[ModelCheckpoint]"
    ) -> str:
        """Returns model name which is string `model_path` appended to `checkpoint_callback.dirpath`."""
        expected_model_path = f"{checkpoint_callback.dirpath}{os.path.sep}"
        if not model_path.startswith(expected_model_path):
            raise ValueError(
                f"{model_path} was expected to start with {expected_model_path}."
            )
        # Remove extension from filepath
        filepath, _ = os.path.splitext(model_path[len(expected_model_path) :])

        return filepath
