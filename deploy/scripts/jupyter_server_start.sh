#!/bin/bash

# Install Jupyter nbextensions

pip install jupyter_contrib_nbextensions

jupyter contrib nbextension install --user

pip install jupyter_nbextensions_configurator

jupyter nbextensions_configurator enable --user


# Start the nb server

start-notebook.sh --NotebookApp.token=''  --NotebookApp.password=''