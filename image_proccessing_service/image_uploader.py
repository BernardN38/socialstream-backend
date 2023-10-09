import pika
from minio import Minio
from io import BytesIO
from PIL import Image
import hashlib
import json
import uuid


class ImageUploader:
    def __init__(self, minio_client):
        self.minio_client = minio_client
        self.connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
        self.channel = self.connection.channel()
    
    def get_image_from_s3(self, image_id):
        img_bytes = self.minio_client.get_object('media-service', image_id)
        img_bytes_io = BytesIO(img_bytes.read())
        return img_bytes_io

    def upload_image_to_s3(self, image_obj, media_id, external_id_full, external_id_compressed, content_type):
        compressed_image_size = image_obj.getbuffer().nbytes
        self.minio_client.put_object('media-service', f'{external_id_compressed}', image_obj, compressed_image_size, part_size=5*1024*1024, content_type=content_type)
        self.publish_message("media_events", "media.compressed", json.dumps({"mediaId": media_id, "externalIdCompressed": f'{external_id_compressed}'}))
        return

    def declare_exchange(self, exchange_name):
        if not self.connection or self.connection.is_closed:
            self.connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
            self.channel = self.connection.channel()
        self.channel.exchange_declare(exchange=exchange_name, exchange_type='direct')
        

    def publish_message(self,exchange_name, routing_key, message):
        if not self.connection or self.connection.is_closed:
            self.connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
            channel = self.connection.channel()
            self.channel = channel
            self.channel.basic_publish(exchange=exchange_name, routing_key=routing_key, body=message)
            return
      
        self.channel.basic_publish(exchange=exchange_name, routing_key=routing_key, body=message)
        print(f" [x] Sent '{message}'")

    def check_if_image_exists(self, hash):
        # Get object information.
        try:
            result = client.stat_object('media-service', hash)
            return True
        except Exception as e:
            print(e)
            return False
