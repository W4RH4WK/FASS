FROM ubuntu:latest

RUN apt update && apt upgrade -y
RUN apt install -y build-essential unzip
RUN gcc --version

RUN useradd -m user
USER user
WORKDIR /home/user
