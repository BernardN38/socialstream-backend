import time
import os
import sys
import threading
from datetime import datetime
import pika
from minio import Minio
import json
from mimetypes import guess_extension
from PIL import Image
import pyheif
from io import BytesIO
import time  
from image_uploader import ImageUploader
import redis




class MediaProcessor:
    def __init__(self, minio_client, rabbitmq_connection, image_uploader, redis_client):
        self.minio_client = minio_client
        self.rabbitmq_connection = rabbitmq_connection
        self.image_uploader = image_uploader
        self.redis_client = redis_client

    def callback(self, ch, method, properties, body):
        start_time = time.time()
        try:
            data = json.loads(body)
            media_id = data["mediaId"]
            external_id_full = data["externalIdFull"]
            external_id_compressed = data["externalIdCompressed"]
            content_type = data["contentType"]
            print("message received by image compressing worker: ", data)
            guess = guess_extension(content_type)
            extension = guess.strip('.')

            image_bytes = self.image_uploader.get_image_from_s3(external_id_full)
            if len(image_bytes.getvalue()) < 1024*1024:
                ch.basic_ack(delivery_tag=method.delivery_tag)
                print("image small enough, compression skipped")
                self.image_uploader.upload_image_to_s3(image_bytes,media_id,external_id_full,external_id_compressed, content_type)
                return
            if extension == "jpg":
                self.compress_upload_jpeg(image_bytes, media_id, external_id_full, external_id_compressed)
            elif extension == "heif":
                self.compress_convert_upload_heic(image_bytes, image_id, external_id_full, external_id_compressed)
            elif extension == "png":
                self.compress_upload_png(image_bytes, image_id, external_id_full, external_id_compressed)
            else:
                print("Extension not recognized:", extension)
            print("Media compressed, media id:", media_id)
        except Exception as e:
            print("Exception:", e)

        end_time = time.time()
        elapsed_time = end_time - start_time
        print(elapsed_time)
        ch.basic_ack(delivery_tag=method.delivery_tag)

    def start_consuming(self, queue_name, exchange_name):
        channel = self.rabbitmq_connection.channel()
        channel.exchange_declare(exchange=exchange_name, exchange_type='topic', durable=True)
        channel.queue_declare(queue=queue_name, durable=True)
        channel.queue_bind(exchange=exchange_name, queue=queue_name, routing_key="media.uploaded")
        channel.basic_qos(prefetch_count=4)
        channel.basic_consume(queue_name, self.callback, auto_ack=False)
        channel.start_consuming()

    def compress_upload_png(self, image_bytes,image_id):
        image = Image.open(image_bytes)
        out = BytesIO()
        resized_image = self.resize_image(image)
        resized_image.quantize(colors=256,method=2)
        resized_image.save(out,
                "png",
                optimize=True,
                quality=75)
        out.seek(0)
        self.image_uploader.upload_image_to_s3(out, image_id, "image/jpeg")
        return


    def compress_upload_jpeg(self, image_bytes, media_id, external_id_full, external_id_compressed):
        image = Image.open(image_bytes)
        out = BytesIO()
        resized_image = self.resize_image(image)
        resized_image.save(out,
                "jpeg",
                optimize=True,
                quality=75)
        out.seek(0)
        self.image_uploader.upload_image_to_s3(out, media_id, external_id_full, external_id_compressed, "image/jpeg")
        out.close()
        return

    def compress_convert_upload_heic(self, image_bytes, image_id):
        image = pyheif.read_heif(image_bytes)
        out = BytesIO()
        pi = Image.frombytes(mode=image.mode, size=image.size, data=image.data)
        resized_image = resize_image(pi)
        self.resized_image.save(out, format="jpeg", optimize=True, quality=75)
        out.seek(0)
        self.image_uploader.upload_image_to_s3(out, image_id, "image/jpeg")
        return

    def resize_image(self, image):
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


def main():
    while True:  # Infinite loop for retrying
        try:
            minio_client = Minio('minio:9000', access_key='minio', secret_key='minio123', secure=False)

            params = pika.ConnectionParameters(host='rabbitmq', heartbeat=200,
                                               blocked_connection_timeout=300)
            rabbitmq_connection = pika.BlockingConnection(params)
            image_uploader = ImageUploader(minio_client)

            # Initialize Redis client
            redis_client = redis.StrictRedis(host='redis', port=6379, db=0)

            media_processor = MediaProcessor(minio_client, rabbitmq_connection, image_uploader, redis_client)

            queue_name = "worker"
            exchange_name = 'media_events'
            print("Listening for RabbitMQ messages")
            media_processor.start_consuming(queue_name, exchange_name)
        except KeyboardInterrupt:
            print('Interrupted')
            try:
                sys.exit(0)
            except SystemExit:
                os._exit(0)
        except Exception as e:
            print(f"Error connecting to RabbitMQ: {e}")
            print("Retrying in 10 seconds...")
            time.sleep(10)  # Wait for 10 seconds before retrying

def create_hash(data):
    sha256_hash = hashlib.sha256()
    sha256_hash.update(data)
    return sha256_hash.hexdigest()

if __name__ == '__main__':
    main()