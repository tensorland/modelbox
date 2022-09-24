#!/bin/bash

# Install Jupyter nbextensions

pip install jupyter_contrib_nbextensions

jupyter contrib nbextension install --user


# Start the nb server

start-notebook.sh --NotebookApp.token=''  --NotebookApp.password=''