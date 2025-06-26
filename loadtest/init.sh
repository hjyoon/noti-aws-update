#!/bin/sh

cd "$(dirname "$0")" || exit

rm -rf venv/ __pycache__/

python -m venv venv

. ./venv/bin/activate

pip install -r ./requirements.txt
