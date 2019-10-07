FROM python:3
MAINTAINER Stefan Safranek, sjsafranek@gmail.com

RUN apt-get update

COPY requirements.txt ./
COPY src ./app/src
COPY Makefile ./app

RUN pip install \
	numpy \
	base58 \
	expiringdict \
	scikit-learn \
	tqdm \
	flask


# RUN pip install -r requirements.txt

CMD cd app && make serve
