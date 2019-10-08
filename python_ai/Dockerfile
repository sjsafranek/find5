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
	tqdm


RUN mkdir /var/lib/findai

CMD cd app/src && python3 server.py --host 0.0.0.0 --port 7005 -D /var/lib/findai
