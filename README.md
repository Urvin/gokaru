# Gokaru storage server

A [golang][golang] lightweight storage and thumbnail server. Storages images for further thumbnailing and any raw files
you want. Thumbnails are creating with Mozjpeg, ZopfliPNG and WebP support.

## Requirements

* [docker][docker]
* docker-compose

## Run

```shell
docker-compose up -d
```

Now Gokaru server runs on 8101 port. Check config.yml for port and other settings

## Usage

### Upload file

Make a put request with body containing file data to /file/{category}/{filename}. You choose category and filename as
you desire. Server responses a 201/Created status in success.

Via curl:

```bash
curl -i http://localhost:8101/file/example/your_first_file --upload-file /path/to/local/file.txt
```

Note that upload method is unprotected, you should set appropriate restrictions on your load balancer / frontend facade
server.

### Download file

Use the GET request with same URL.

```bash
wget http://localhost:8101/file/example/your_first_file
```

### Delete file

Use the DELETE request with same URL.

```bash
curl -i -X DELETE http://localhost:8101/file/example/your_first_file
```

Note that delete method is unprotected, you should set appropriate restrictions on your load balancer / frontend facade
server.

### Upload image

Make a put request with body containing image data to /image/{category}/{filename}. You choose category and filename as
you desire. Server responses a 201/Created status in success.

Via curl:

```bash
curl -i http://localhost:8101/image/example/your_first_image --upload-file /path/to/local/image.png
```

Note that upload method is unprotected, you should set appropriate restrictions on your load balancer / frontend facade
server.

### Download image origin

Use the GET request with same URL.

```bash
wget http://localhost:8101/image/example/your_first_image
```

### Delete image

Use the DELETE request with same URL.

```bash
curl -i -X DELETE http://localhost:8101/image/example/your_first_image
```

Note that delete method is unprotected, you should set appropriate restrictions on your load balancer / frontend facade
server.

### Thumbnail image

**Define your image width and height**
If you want to get an image with 200px width and no matter height, set height to 0. Thumbnailer will calculate it using
original image aspect ratio. The same works with width. If you set to 0 both params, the source width and height will be
used.

**Define cast flag**
The resulting cast flag should be an integer, obtained via bitwise OR among available cast flags.

**Define format**
Gokaru accepts PNG, GIF, WEBP and JPG output. If the origin image is animated and output format supports animation,
output would be also animated. Also, Gokaru sends WEBP format to browser accepting image/webp regardless your extension.

**Calculate security signature**

Use md5 or murmur signature, according to config.yml.
Both signature algorithms use your own salt, described in config.yml.

***MD5 algorithm***

Signature is a md5 hash of concatenated stings of salt, width, height, cast, category and filename without extension

```bash
**_# salt = secretsalt
# source_type = image
# category = example
# filename = your_first_image.jpg
# with = 100
# height = 200
# cast = 8_**
echo -n secretsalt/image/example/your_first_image.jpg/100/200/8 | md5sum
```

***MurMur3 algorithm***

Signature is a MurMur3 32-based hash in of concatenated stings of salt, width, height, cast, category and filename without extension
MumMur3 is shorter than MD5, that is better for URLs, and is a default Gokaru signature algorithm

```php
<?php
    // composer require lastguest/murmurhash
    use lastguest\Murmur;

    $salt = 'secretsalt';
    $sourceType = 'image';
    $category = 'example';
    $fileName = 'your_first_image.jpg';
    $with = 100;
    $height = 200;
    $cast = 8;


    echo Murmur::hash3(
        $salt . '/' .
        $sourceType . '/' .
        $category . '/' .
        $fileName . '/' .
        (string)$with . '/' .
        (string)$height . '/' .
        (string)$cast
    );
```

**Combine parts of your URL**

Request /source_type/signature/category/width/height/cast/filename.extension one

```bash
#md5
wget http://localhost:8101/image/3ac8ee6f420b812ec95176bbb54d7653/example/100/200/8/your_first_image.jpg
```
```bash
#murmur
wget http://localhost:8101/image/1d7ulp9/example/100/200/8/your_first_image.jpg
```

### Cast flags

- _CAST_RESIZE_TENSILE = 2_ - stretch image directly into defined width and height ignoring aspect ratio
- _CAST_RESIZE_PRECISE = 4_ - keep aspect-ratio, use higher dimension
- _CAST_RESIZE_INVERSE = 8_ - keep aspect-ratio, use lower dimension
- _CAST_TRIM = 16_ - remove any edges that are exactly the same color as the corner pixels
- _CAST_EXTENT = 32_ - set output canvas exactly defined width and height after image resize
- _CAST_OPAGUE_BACKGROUND = 64_ - set image white opaque background
- _CAST_TRANSPARENT_BACKGROUND = 128_ - create a transparent background for an image
- _CAST_TRIM_PADDING = 265_ - Adds 10px (or other, according to config.yml) padding around your trimmed image

## Clients

- [Gokaru-php-client][phikaru] PHP client

## Author

Yuriy Gorbachev <yuriy@gorbachev.rocks>

## License

This module is licensed under the [GLWTPL][license] license.

[golang]:<https://golang.org/>

[docker]:<https://www.docker.com/>

[license]:<https://github.com/me-shaon/GLWTPL>

[phikaru]:<https://github.com/Urvin/gokaru-php-client>
