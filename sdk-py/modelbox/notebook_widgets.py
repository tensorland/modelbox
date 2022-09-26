from email import header
from importlib.metadata import metadata
from modelbox.modelbox import Experiment

from tabulate import tabulate
from IPython.display import Markdown

import matplotlib.pyplot as plt


class ExperimentDisplay:
    def __init__(self, experiment: Experiment) -> None:
        self._experiment = experiment

    def info(self):
        experiment_info = [
            ["id", self._experiment.id],
            ["name", self._experiment.name],
            ["owner", self._experiment.owner],
            ["namespace", self._experiment.namespace],
            ["creation time", self._experiment.created_at],
            ["updated time", self._experiment.updated_at],
        ]
        return Markdown(
            tabulate(experiment_info, headers=["Experiment", ""], tablefmt="github")
            + "\n"
            + "##### Metadata"
            + "\n"
            + tabulate(
                self._experiment.metadata().metadata.items(),
                headers=["", ""],
                tablefmt="github",
            )
        )

    def metrics(self):
        all_metrics = self._experiment.all_metrics()
        fig, axs = plt.subplots(len(all_metrics.keys()))
        index = 0
        for key, metrics in all_metrics.items():
            m_values = [(mv.step, mv.value) for mv in metrics]
            axs[index].plot(*zip(*m_values))
            index = index + 1
        plt.show()


    def events(self):
        events = self._experiment.events()
        events_table =[]
        for event in events:
            events_table.append((event.wallclock_time, event.name, event.source.name))
        return Markdown(tabulate(events_table, headers=["wallclock", "event", "source"], tablefmt="github"))


