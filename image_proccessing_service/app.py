import pika
import time
import os
from io import BytesIO
from PIL import Image
import sys
from s3_helpers import upload_image_to_s3, get_image_from_s3
from mimetypes import guess_extension
import threading
from datetime import datetime
from minio import Minio
import json
import pyheif

minio_client = Minio('minio:9000',
                     access_key='minio',
                     secret_key='minio123',
                     secure=False)


def callback(ch, method, properties, body):
    start_time = time.time()
    try:
        data = json.loads(body)
        image_id = data["mediaId"]
        content_type = data["contentType"]

        guess = guess_extension(content_type)
        extension = guess.strip('.')

        image_bytes = get_image_from_s3(image_id)
        if image_bytes.__sizeof__() < 1 << 20:
            ch.basic_ack(delivery_tag=method.delivery_tag)
            return
        match extension:
            case "jpg":
                compress_upload_jpeg(image_bytes, image_id)
            case "heif":
                compress_convert_upload_heic(image_bytes, image_id)
            case "png":
                compress_upload_png(image_bytes,image_id)
            case __:
                print("extension not recognized", extension)
        print("media compressed, media id: ", image_id)
    except Exception as e:
        print("exception", e)
    
    end_time = time.time()
    elapsed_time = end_time - start_time
    print(elapsed_time)
    ch.basic_ack(delivery_tag=method.delivery_tag)


def compress_upload_png(image_bytes,image_id):
    image = Image.open(image_bytes)
    out = BytesIO()
    resized_image = resize_image(image)
    resized_image.quantize(colors=256,method=2)
    resized_image.save(out,
               "png",
               optimize=True,
               quality=75)
    out.seek(0)
    upload_image_to_s3(out, image_id, "image/png")


def compress_upload_jpeg(image_bytes, image_id):
    image = Image.open(image_bytes)
    out = BytesIO()
    resized_image = resize_image(image)
    resized_image.save(out,
               "jpeg",
               optimize=True,
               quality=75)
    out.seek(0)
    upload_image_to_s3(out, image_id, "image/jpeg")
    return

def compress_convert_upload_heic(image_bytes, image_id):
    image = pyheif.read_heif(image_bytes)
    out = BytesIO()
    pi = Image.frombytes(mode=image.mode, size=image.size, data=image.data)
    resized_image = resize_image(pi)
    resized_image.save(out, format="jpeg", optimize=True, quality=75)
    out.seek(0)
    upload_image_to_s3(out, image_id, "image/jpeg")
    return

def resize_image(image):
    if image.width > 1920 or image.height > 1080:
            new_width, new_height = 1920, 1080
            aspect_ratio = image.width / image.height
            if aspect_ratio > 1.777:  # check if aspect ratio is wider than 16:9
                new_width = int(new_height * aspect_ratio)
            else:
             new_height = int(new_width / aspect_ratio)
            # Resize the image
            image = image.resize((new_width, new_height))
    return image

def worker(queue_name):
    connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
    channel = connection.channel()

    # Declare the exchange
    exchange_name = 'media_events'
    channel.exchange_declare(exchange=exchange_name, exchange_type='topic')

    channel.queue_declare(queue=queue_name, durable=True)  
    channel.queue_bind(exchange=exchange_name, queue=queue_name, routing_key="media.uploaded")

    channel.basic_qos(prefetch_count=2)
    channel.basic_consume(queue_name, callback, auto_ack=False)
    channel.start_consuming()


def main():
    time.sleep(10)
    for i in range(10):
        t = threading.Thread(target=worker, args=("image-proccessing",))
        t.start()
    t.join()


if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print('Interrupted')
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)