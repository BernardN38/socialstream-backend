from minio import Minio
from io import BytesIO
from PIL import Image

minio_client = Minio('minio:9000',
                  access_key='minio',
                  secret_key='minio123',
                  secure=False)


# def get_image_from_s3(image_id):
#     img_bytes = minio_client.get_object(
#             'media-service-socialsphere1', image_id)
#     img_bytes_io = BytesIO(img_bytes.read())
#     image = Image.open(img_bytes_io)
#     return image

def get_image_from_s3(image_id):
    img_bytes = minio_client.get_object(
            'media-service', image_id)
    img_bytes_io = BytesIO(img_bytes.read())
    return img_bytes_io

def upload_image_to_s3(image_obj, image_id, content_type):
    # bucket.upload_fileobj(image_obj, image_id)  
    minio_client.put_object('media-service', image_id, image_obj, -1, part_size=5*1024*1024, content_type=content_type )
    return 