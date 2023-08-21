FROM python:slim-buster

WORKDIR /app

# RUN apt-get update && apt-get install -y imagemagick

COPY requirements.txt requirements.txt
RUN pip3 install  -r requirements.txt

COPY . .

CMD [ "python3", "-u", "app.py"]